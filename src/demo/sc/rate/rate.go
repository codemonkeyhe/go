// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rate provides a rate limiter.
package rate

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

/* Limit
令牌桶的生产速率  单位是 每秒多少个
例如 每秒10个  limit=10  相当于每100毫秒生成1个token
limit的速度 *  单位时间  即可得 单位时间 生成的token数目,即 tokensFromDuration 方法
*/

// Limit defines the maximum frequency of some events.
// Limit is represented as number of events per second.
// A zero Limit allows no events.
type Limit float64

// durationFromTokens is a unit conversion function from the number of tokens to the duration
// of time it takes to accumulate them at a rate of limit tokens per second.
// 当limit为0时，并不会发生除0报错，因为是浮点数，不是整数除0
// 参数token必须>=0
func (limit Limit) durationFromTokens(tokens float64) time.Duration {
	seconds := tokens / float64(limit)
	return time.Nanosecond * time.Duration(1e9*seconds)
}

// tokensFromDuration is a unit conversion function from a time duration to the number of tokens
// which could be accumulated during that duration at a rate of limit tokens per second.
// 把秒拆分成 秒的整数秒和秒的纳秒两部分，是为了精度
func (limit Limit) tokensFromDuration(d time.Duration) float64 {
	// Split the integer and fractional parts ourself to minimize rounding errors.
	// See golang.org/issues/34861.
	sec := float64(d/time.Second) * float64(limit)
	nsec := float64(d%time.Second) * float64(limit)
	return sec + nsec/1e9
}

// Inf is the infinite rate limit; it allows all events (even if burst is zero).
const Inf = Limit(math.MaxFloat64)

// Every converts a minimum time interval between events to a Limit.
// 通过每interval生成1个token 来反算limit， 例如100毫秒生成1个token，反算limit=10
func Every(interval time.Duration) Limit {
	if interval <= 0 {
		return Inf
	}
	return 1 / Limit(interval.Seconds())
}

/*
不是每次调用Allow,Wait, Reserve都必然消费1个token， 只有调用成功时，即reserveN().ok == true才算消费1个token
allow 用于 限速时直接拒绝或者丢弃场景
wait 用于 限速时的等待场景，允许通过ctx设置 上游的等待时长，超时时则触发cancel归还token.
wait 会自动帮我们sleep，超时会自动cancel
reserve 用于 限速时返回等待时长，reserve没有超时设置，由上游业务自己sleep，而且是必须sleep
reserve和wait的区别在于， reserve 把sleep和cancel的维护逻辑 丢给了 业务调用方去管理，让调用方更加灵活
而wait则封装好了 sleep和cancel的逻辑， 两者对外围代码开放的粒度不同
limiter可以用作单个进程下单个接口 或者 单个API层面的 限流器，用作单个服务层面的限流则不合适
多服务多进程的分布式限流，limiter则不适用，需要把分布式的流量汇总计算
*/

// A Limiter controls how frequently events are allowed to happen.
// It implements a "token bucket" of size b, initially full and refilled
// at rate r tokens per second.
// Informally, in any large enough time interval, the Limiter limits the
// rate to r tokens per second, with a maximum burst size of b events.
// As a special case, if r == Inf (the infinite rate), b is ignored.
// See https://en.wikipedia.org/wiki/Token_bucket for more about token buckets.
//
// The zero value is a valid Limiter, but it will reject all events.
// Use NewLimiter to create non-zero Limiters.
//
// Limiter has three main methods, Allow, Reserve, and Wait.
// Most callers should use Wait.
//
// Each of the three methods consumes a single token.
// They differ in their behavior when no token is available.
// If no token is available, Allow returns false.
// If no token is available, Reserve returns a reservation for a future token
// and the amount of time the caller must wait before using it.
// If no token is available, Wait blocks until one can be obtained
// or its associated context.Context is canceled.
//
// The methods AllowN, ReserveN, and WaitN consume n tokens.
type Limiter struct {
	mu    sync.Mutex
	limit Limit //令牌桶的生成速率，每1秒生成limit个token
	burst int   // 令牌桶的大小

	// tokens可以是正数或者0 或者 负数
	// tokens为正数时，代表当前有token可用，
	//	如果申请的数目n <= token,则会申请成功
	//	如果申请的数目n > token,则会申请成功,token会变成负数而已
	// tokens为负数时，代表着已经预支了 未来时间段生成的token，
	// 接下来所有的token申请，都不会立即生效，而是会在更遥远的未来才能生效，生效的时间由timeToAct来决定
	tokens float64 //当前桶剩余的token数目

	// last is the last time the limiter's tokens field was updated
	last time.Time // token的更新时间，即最后一次 分配token的时间
	// lastEvent is the latest time of a rate-limited event (past or future)
	// lastEvent只有1个作用，就是用于CancelAt时，准确计算应该归还多少个token
	lastEvent time.Time // 代表最后一次 reservation 的timeToAct
}

// Limit returns the maximum overall event rate.
func (lim *Limiter) Limit() Limit {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	return lim.limit
}

/*
 令牌桶的大小，令牌桶的优点之一 就是允许突发流量，burst本意就是 突发，
 burst 限制了一次申请的token数目，例如burst=10 ，则业务一次最多申请10个token
 burst 意味着 流量突发场景下的波峰高度，burst=100 ，即一开始最多同时允许100个流量通过
 如果，token生产速度大于消费速度的情况下， 则token会增加，最多不会超过burst，为下一次流量突发场景做准备
 如果，token生产速度小于消费速度的情况下， 则token会减少为0或者负数，此时也允许一次性来burst个流量，只是token会变成负值而已
 任何场景下，都不允许瞬时的流量大于burst，否则违背了令牌桶的初衷
 这一点也是 CancelAt的 restoreTokens 计算核心之一

 如果初始的burst个token已经被消耗完了,lim.tokens=0, 接下来 消费速度 等于 生产速度，那么 limit的生产速率 就是 业务代码的 QPS限制了。 如何确定 limit的大小，应该是根据业务场景的消费速度 来反推生产速度。
 比如某个API的QPS是100，那么limit=100
*/

// Burst returns the maximum burst size. Burst is the maximum number of tokens
// that can be consumed in a single call to Allow, Reserve, or Wait, so higher
// Burst values allow more events to happen at once.
// A zero Burst allows no events, unless limit == Inf.
func (lim *Limiter) Burst() int {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	return lim.burst
}

// NewLimiter returns a new Limiter that allows events up to rate r and permits
// bursts of at most b tokens.
func NewLimiter(r Limit, b int) *Limiter {
	return &Limiter{
		limit: r,
		burst: b,
	}
}

// Allow is shorthand for AllowN(time.Now(), 1).
// 当前时间 是否允许 1个流量 通过
// 只有return为true时，才算消耗了1个token，
// return为false时，没有消费token，业务代码则应该忽略或者丢弃该流量
func (lim *Limiter) Allow() bool {
	return lim.AllowN(time.Now(), 1)
}

// AllowN reports whether n events may happen at time now.
// Use this method if you intend to drop / skip events that exceed the rate limit.
// Otherwise use Reserve or Wait.
// 当前时间 是否允许 N个流量 通过
// 当return=false时，意味着限流了，此时业务代码 应该 drop/skip 本次流量
func (lim *Limiter) AllowN(now time.Time, n int) bool {
	return lim.reserveN(now, n, 0).ok
}

/*
Reservation
 最核心的event对象， 每一次allow, wait, reserve 都会生成一个新的reservation对象
 整个Reservation本质上是为了 CancelAt这个方法做准备, 这样把很多事情搞复杂了，
 真正业务上，一般的限速就是skip or drop or  wait， 这种允许wait了一段时间再cancel场景，相当于泼出去的水，泼了一半还要收回来，增加了复杂度。
 假设不支持Cancel场景，Reservation只需要2个字段即可： ok和 timeToAct

 Reservation有3个时间点
1 成功申请时间点  r.applyTime
2 最终执行之间点  r.timeToAct
3 从前一个prevR.timeToAct到自己的r.timeToAct的时长 r.dur= r.limit.durationFromTokens(r.tokens)
满足关系  prevR.timeToAct + r.dur == r.timeToAct
把r.dur 记作 分割时长

r 的实际等待时长  r.waitTime = r.timeToAct-r.applyTime
r.waitTime 包含了 这区间 所有的 r0.dur r1.dur ... r.dur

可以理解为，lim.tokens在prevR.timeToAct刚刚恢复到0，
然后又被预支了r.tokens, 即lim.tokens=lim.tokens-r.tokens
因此本次的r需要持续等待r.dur时长，让lim.tokens在r.timeToAct恢复到0


以时间线作为X轴，lim.tokens为Y轴，lim.tokens顺着X轴往右增长，直到桶的上限burst为止
Y轴lim.tokens每次下降，意味着发生了一次tokens分配
lim.tokens>0的区间，且满足 lim.tokens> r.tokens 则本次r可以立即执行，否则,r将延迟执行
lim.tokens<0的区间，未来每一次r的申请 只会更加延迟。

r的意义在于 锁定了X轴的[prevR.timeToAct, r.timeToAct]这个时间段,且限定了流量为r.tokens,
r.tokens兑现时间为r.timeToAct
1 在同一个时间段内，不应该存在其他立即执行的r, 这些r满足waitTime=0, applyTime=timeToAct,
因为假设存在立即执行的r, 意味着tokens够用，那本次的r就不应该阻塞等待。除非有其他的r提前归还了tokens

2 在同一个时间段内，绝对不存在 延迟执行的r1, 或者说 X轴上延迟执行的r1.[r1.prevR.timeToAct, r1.TimeToAct]不会与r的时间段[r.prevR.timeToAct, r.timeToAct]有交叉. 把X轴切分成一段段的，每一段分给每个r,分割区间的长度即 分割时长r.dur。

*/

// A Reservation holds information about events that are permitted by a Limiter to happen after a delay.
// A Reservation may be canceled, which may enable the Limiter to permit additional events.
type Reservation struct {
	// reserveN ok := n <= lim.burst && waitDuration <= maxFutureReserve
	ok bool // 是否成功申请token

	lim *Limiter //指向token池子，主要用于CancelAt时 获取token池

	// 本次申请到的token数目，或者说预分配的数目，通常是外围的函数参数n
	tokens int

	// 如果token池够用,timeToAct=now
	// 如果不够用， timeToAct表示未来某个时间点token池子 才满足 本次申请tokens数目, 因此, 业务必须到timeToAct才能放过流量
	// 在timeToAct之前，要么drop/skip，即allow方法，要么阻塞等待，即wait方法
	timeToAct time.Time

	// This is the Limit at reservation time, it can change later.
	//记录当时申请tokens时的 生产速率，因为速率会变
	// 只有1个作用，用于CancelAt时，计算 单位时间内 对应的token数目
	limit Limit
}

// OK returns whether the limiter can provide the requested number of tokens
// within the maximum wait time.  If OK is false, Delay returns InfDuration, and
// Cancel does nothing.
func (r *Reservation) OK() bool {
	return r.ok
}

// Delay is shorthand for DelayFrom(time.Now()).
func (r *Reservation) Delay() time.Duration {
	return r.DelayFrom(time.Now())
}

// InfDuration is the duration returned by Delay when a Reservation is not OK.
const InfDuration = time.Duration(1<<63 - 1)

// DelayFrom returns the duration for which the reservation holder must wait
// before taking the reserved action.  Zero duration means act immediately.
// InfDuration means the limiter cannot grant the tokens requested in this
// Reservation within the maximum wait time.
// 计算必须要等待的时长=timeToAct-now
func (r *Reservation) DelayFrom(now time.Time) time.Duration {
	if !r.ok {
		return InfDuration
	}
	delay := r.timeToAct.Sub(now)
	if delay < 0 {
		return 0
	}
	return delay
}

// Cancel is shorthand for CancelAt(time.Now()).
func (r *Reservation) Cancel() {
	r.CancelAt(time.Now())
	return
}

/*
 最最难理解的地方 restoreTokens的计算逻辑 修改了一堆UT才搞明白，果然是实践出真知
上游调用CancelAt，其实并不知道 是否真的 Cancel成功了，没有返回码告知上游Cancel结果
不是每次cancel都一定能归还r.tokens, 能消除所有的影响，只是尽力而为，as much as possible而已

因此，我们是尽力归还，不是一定会归还。


Cancel场景
Cancel只用于延迟执行的r, 立即执行的r.timeToAct==r.applyTime,再等到调用Cancel时，cancelTime必然在r.timeToAct之后了，不满足释放条件了
也就是 那些调用AllowN的函数，不可能调用Cancel 只有调用ReserveN和WaitN返回的reversition对象，且对象不是立即执行的，才有Cancel的可能


在 [applyTime, timeToAct]这个时间段内,才能正常执行 r.CancelAt. 分成2种情况:
1 简单情况： 在[r.ApplyTime, r.CancelTime]，没有其他r的申请，则 r归还的tokens就是restoreTokens=r.tokens
2 复杂情况:  在[r.ApplyTime, r.CancelTime]，发生了r1,r2..rn的申请，lastEvent=rn.TimeToAct
则 restoreTokens  = r.tokens - r.limit.durationFromTokens(rn.TimeToAct - r.TimeToAct)
if restoreTokens <=0 则不用归还
其中：
r.limit.durationFromTokens(rn.TimeToAct - r.TimeToAct) 代表着 r1~rn申请的tokens总和tokensSum
相当于在 r.TimeToAct之后，要从r.tokens预先扣除 tokensSum 留给r1~rn使用

问题来了，为什么计算restoreTokens要扣除tokensSum?
主要原因：burst,  避免瞬时流量大于burst
 r.tokens的归还会增加当前lim.tokens，但是不应该影响r1~rn的timeToAct
会影响下一次新分配rk的timeToAct, rk.timeToAct可能会与r1~rn.timeToAct重叠

在不归还的前提下， 本次r.tokens占有区间是[prevR.timeToAct, r.timeToAct]这个时间段,
本来在r.timeToAct时，lim.tokens恢复到0.  因为提前归还，导致在r.timeToAct有一定的restoreToken

tokensSum = r1.tokens + r2.tokens + ... rn.tokens
在[r.timeToAct,  rn.TimeToAct]时间段,  已经产生了 tokensSum，
显而易见，tokensSum 大概率大于burst，但是分摊到 [r.timeToAct,  rn.TimeToAct]时间段内的若干次 reservation对象，能保证任意时刻的流量不超过burst。
任意时间点的瞬时流量都不能超过桶的大小burst

r.prevR.timeToAct
^
|     r.timeToAct                                         rn.timeToAct
|        |                                                 |
|--------|---------|---------|---------|---------|---------|
r.tokens  r1.tokens r2.tokens  ......    ......   rn.tokens

换个角度思考, 本次r.tokens占有区间是[prevR.timeToAct, r.timeToAct]这个时间段, 如果归还了r.tokens,
那么r.tokens必然会在其他的时间段生效， 其他时间段在分配tokens时，满足了瞬时流量不超过burst的限制，
那么 不能因为 突然归还r.tokens，便突破了burst的限制。
主要是没有一个账本记录所有分配出去的reservation对象，根本不清楚某一个时间段到底有多少个reservation执行timeToAct，唯一明白的就是，那个时间段的tokens消耗必然不超过burst。每次分配的tokens都在未来时间轴上占据了一定的时间片段。

在不归还的前提下， 本次r.tokens占有区间是[prevR.timeToAct, r.timeToAct]这个时间段, 这个时间段内必然不存在其他的延迟执行的reservation对象
本次r.tokens是在r.timeToAct去兑现的， 既然要归还r.tokens，那这r.tokens的兑现时间必然在r.PrevR.timeToAct之后，也可能在 rn.TimeToAct之后才兑现,取决于新的reversation对象的申请时间

因为归还了r.tokens, 那么自r.timeToAct到rn.TimeToAct，把r.tokens给任意时间段都可能导致流量超过burst。
在[r.timeToAct,  rn.TimeToAct]时间段的 任意时刻都会发生新的reservation申请，
	不归还时，新分配的reservation对象的tokens兑现时间必然在 rn.timeToAct之后，也不会超过burst大小
	因为归还了r.tokens，导致lim.tokens增加了，后续新分配的rm.tokens的rm.timeToAct 会与 r1~rn.TimeToAct重叠，导致违背了burst的限制。


因此，需要找到一种尽力归还的简单办法，同时不违背瞬时流量不超过burst的原则

如果自r.timeToAct已经发生了tokensSum的分配，

如果tokensSum >= r.tokens, 则不用归还了
	还回去，则意味着 r.tokens可能发生于 r1~rn的timeToAct, r.tokens与rk.tokens的叠加，有大于burst的风险，所以不用归还

如果tokensSum < r.tokens, 则尽力归还 r.tokens-tokensSum
	意味着 [r.timeToAct, rn.timeToAct]生产的tokens
	从tokensSum变成了 tokensSum + (r.tokens-tokensSum) = r.tokens
	已知r.tokens必然小于burst，因此，避免了瞬时流量违背burst的风险

关联问题
Q1 既然这样归还是为了避免瞬时流量违背burst，那么初始化时，设置burst为无限大是否可行?
从代码设计角度，burst定义为int，最小值是0，最大值就是int的上边界，不可能是一个无限大的场景，初始化的API必须要提供一个burst的大小。 只要设定了burst的大小，就会存在上面的问题。

Q2 有没有其他的归还策略？更符合直觉的归还策略？
更符合直觉的策略:
策略1: 归还r的时候，依然归还r.tokens,然后把r1~rn的timeToAct都修改一边，往前移动一点。
这个显然无法实现，r1~rn都是泼出去的水了，可能有业务代码已经获取到tokens，正在sleep了，等到timeToAct去兑现申请的tokens

策略2: 归还r的时候，依然归还r.tokens ，为了避免超过burst，把r.tokens放到rn.timeToAct之后才允许分配。
这个也显然难以实现，复杂度巨大。

策略3: 归还r的时候，依然归还r.tokens ，限制这r.tokens只能在 [r.prevR.timeToAct, r.timeToAct]去使用
违背逻辑，既然这样，那为啥最初要归还呢 其次,依然实现复杂


最好的办法就是， 压根不支持Cancel，会让问题简单很多


回归到算法实现细节，假设每次r.Cancel都归还r.tokens，不用扣除预支的tokens，代码如下所示：
	restoreTokens := r.tokens
	// advance time to now
	// 从lim.last到归还时间点now时，令牌桶的预期tokens数目
	now, _, tokens := r.lim.advance(now)
	// calculate new number of tokens
	tokens += restoreTokens
	if burst := float64(r.lim.burst); tokens > burst {
		tokens = burst
	}
	r.lim.last = now
	r.lim.tokens = tokens

在这里，我们的确规避了tokens总数不超过lim.burst,
但是，实际情况是，已经预支了tokensSum出去了，这里的tokens早已经是负值，永远都小于burst。
这里对瞬时流量不超过burst的限制并不起任何作用。

参照 TestCancelTokenN 的例子，
如果每次都归还r.tokens，
则r在t999归还 1000个tokens，
restoreTokens=1000
tokens=-1002 +999+ 1000 = 997 而非原来的995
t999时刻的lim.tokens=997,
到了t1002时刻，lim.tokens=997+3=1000
t1002时刻本来就有r1.timeToAct的2个tokens，最终实质发生的tokens总数为1002 > 1000的burst
*/

// CancelAt indicates that the reservation holder will not perform the reserved action
// and reverses the effects of this Reservation on the rate limit as much as possible,
// considering that other reservations may have already been made.
func (r *Reservation) CancelAt(now time.Time) {
	if !r.ok {
		// 都没申请成功，那就没必要cancel
		return
	}

	r.lim.mu.Lock()
	defer r.lim.mu.Unlock()

	if r.lim.limit == Inf || r.tokens == 0 || r.timeToAct.Before(now) {
		// 速率是无限的，也没必要返还token
		// 本次申请持有的tokens为0，也不用返还
		// 返还时间点now 在 本次 event的timeToAct之后了，不允许返还
		// 因为过了timeToAct之后，生米煮成熟饭了，外围业务代码已经付出了足够的阻塞成本了,
		// tokens池子也已经被本次event持有的tokens 持续占有很长一段时间了, 无法返还
		return
	}

	// calculate tokens to restore
	// The duration between lim.lastEvent and r.timeToAct tells us how many tokens were reserved
	// after r was obtained. These tokens should not be restored.
	// r.timeToAct 是满足r.tokens的时间点，这个时间点必然>=now
	// r.lim.lastEvent 表明 自动 本次r的申请token成功后， 可能会有其他的reservation对象也申请了tokens，
	// lastEvent指向了 最后一次 申请tokens的reservation对象的timeToAct.
	// tokensFromDuration( r.lim.lastEvent - r.timeToAct ) 是为了表明 本次r申请成功后，发生了若干次其他的申请，那么 这期间总共预分配了多少个tokens，这些tokens必须从restoreTokens中扣除
	// 否则未来会发生 瞬时QPS大于burst的场景，违背了令牌桶的基本原理 可以参照 TestCancelTokenN 去理解
	restoreTokens := float64(r.tokens) - r.limit.tokensFromDuration(r.lim.lastEvent.Sub(r.timeToAct))

	if false {
		// for  TestCancelTokenN , simulate QPS > burst
		restoreTokens = float64(r.tokens)
	}
	if restoreTokens <= 0 {
		return
	}
	// advance time to now
	// 从lim.last到归还时间点now时，令牌桶的预期tokens数目
	now, _, tokens := r.lim.advance(now)
	// calculate new number of tokens
	tokens += restoreTokens
	if burst := float64(r.lim.burst); tokens > burst {
		tokens = burst
	}
	// update state
	r.lim.last = now
	r.lim.tokens = tokens
	// 在r申请token之后，一直没有其他的token申请，本次的r就是最后1次申请，则需要回溯上一次的lastEvent
	// 通过 r.timeToAct - r.tokens占用的时间  = 上一次其他r的timeToAct, 然后修正lastEvent的指向
	if r.timeToAct == r.lim.lastEvent {
		prevEvent := r.timeToAct.Add(r.limit.durationFromTokens(float64(-r.tokens)))
		if !prevEvent.Before(now) {
			r.lim.lastEvent = prevEvent
		}
	}

	return
}

// Reserve is shorthand for ReserveN(time.Now(), 1).
func (lim *Limiter) Reserve() *Reservation {
	return lim.ReserveN(time.Now(), 1)
}

// ReserveN returns a Reservation that indicates how long the caller must wait before n events happen.
// The Limiter takes this Reservation into account when allowing future events.
// The returned Reservation’s OK() method returns false if n exceeds the Limiter's burst size.
// Usage example:
//   r := lim.ReserveN(time.Now(), 1)
//   if !r.OK() {
//     // Not allowed to act! Did you remember to set lim.burst to be > 0 ?
//     return
//   }
//   time.Sleep(r.Delay())
//   Act()
// Use this method if you wish to wait and slow down in accordance with the rate limit without dropping events.
// If you need to respect a deadline or cancel the delay, use Wait instead.
// To drop or skip events exceeding rate limit, use Allow instead.
// 调用方must wait, 即必须调用 time.Sleep(r.Delay())
func (lim *Limiter) ReserveN(now time.Time, n int) *Reservation {
	r := lim.reserveN(now, n, InfDuration)
	return &r
}

// Wait is shorthand for WaitN(ctx, 1).
func (lim *Limiter) Wait(ctx context.Context) (err error) {
	return lim.WaitN(ctx, 1)
}

// WaitN blocks until lim permits n events to happen.
// It returns an error if n exceeds the Limiter's burst size, the Context is
// canceled, or the expected wait time exceeds the Context's Deadline.
// The burst limit is ignored if the rate limit is Inf.
// 如果ctx被cenceled或者超出ctx的deadline， 则会返回ctx.Err()，如果已经分配到一个新的reservation,则会归还已经获得的token
func (lim *Limiter) WaitN(ctx context.Context, n int) (err error) {
	lim.mu.Lock()
	// 拷贝出来，减少锁的占用时长
	burst := lim.burst
	limit := lim.limit
	lim.mu.Unlock()
	// 申请数目n大于令牌桶大小，则报错， 符合预期，即突发流量不得超过令牌桶大小
	if n > burst && limit != Inf {
		return fmt.Errorf("rate: Wait(n=%d) exceeds limiter's burst %d", n, burst)
	}
	// Check if ctx is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Determine wait limit
	now := time.Now()
	// 最长等待时间，由ctx设置  waitLimit=deadLine-now
	waitLimit := InfDuration
	if deadline, ok := ctx.Deadline(); ok {
		waitLimit = deadline.Sub(now)
	}
	// Reserve
	r := lim.reserveN(now, n, waitLimit)
	if !r.ok {
		// 申请失败， 这里不可能是 n >burst，因为前面已经排除掉了
		// 这里的ok=false只有一种原因，即满足n个token的等待时长waitDuration > waitLimit，即超过了等待限制
		return fmt.Errorf("rate: Wait(n=%d) would exceed context deadline", n)
	}
	// Wait if necessary
	// delay = r.TimeToAct - now
	delay := r.DelayFrom(now)
	if delay == 0 { // 说明当前token池满足本次n个token的申请，因此timeToAct==now,delay=0
		return nil
	}
	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-t.C:
		// We can proceed.
		return nil
	case <-ctx.Done():
		// Context was canceled before we could proceed.  Cancel the
		// reservation, which may permit other events to proceed sooner.
		r.Cancel()
		return ctx.Err()
	}
}

// SetLimit is shorthand for SetLimitAt(time.Now(), newLimit).
func (lim *Limiter) SetLimit(newLimit Limit) {
	lim.SetLimitAt(time.Now(), newLimit)
}

// SetLimitAt sets a new Limit for the limiter. The new Limit, and Burst, may be violated
// or underutilized by those which reserved (using Reserve or Wait) but did not yet act
// before SetLimitAt was called.
func (lim *Limiter) SetLimitAt(now time.Time, newLimit Limit) {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	// 必须先把时间线 push到 now时刻，对其now时刻的tokens数目，再修正limit速率，才是合理的
	now, _, tokens := lim.advance(now)

	lim.last = now
	lim.tokens = tokens
	lim.limit = newLimit
}

// SetBurst is shorthand for SetBurstAt(time.Now(), newBurst).
func (lim *Limiter) SetBurst(newBurst int) {
	lim.SetBurstAt(time.Now(), newBurst)
}

// SetBurstAt sets a new burst size for the limiter.
func (lim *Limiter) SetBurstAt(now time.Time, newBurst int) {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	// 必须先把时间线 push到 now时刻，对其now时刻的tokens数目，再修正burst才是合理的
	// 理论上，可以有个 SetLimitAndburst的方法
	now, _, tokens := lim.advance(now)

	lim.last = now
	lim.tokens = tokens
	lim.burst = newBurst
}

// reserveN is a helper method for AllowN, ReserveN, and WaitN.
// maxFutureReserve specifies the maximum reservation wait duration allowed.
// reserveN returns Reservation, not *Reservation, to avoid allocation in AllowN and WaitN.  // 最核心方法,
// 当lim.tokens>=n 时，能申请成功，lim.tokens-=n且 r.TimeToAct = now
// 当lim.tokens<n 时，也能申请成功，lim.tokens-=n， lim.tokens<0,且 r.TimeToAct = future,  等到时间点到future时，必有lim.tokens增长到0
// maxFutureReserve 参数是单独给 WaitN使用的
// ReserveN 的maxFutureReserve 为 InfDuration ，几乎是无限的
// AllowN 的 maxFutureReserve 为 0  那么只要tokens<0，则必然有waitDuration>maxFutureReserve，则 ok=false
func (lim *Limiter) reserveN(now time.Time, n int, maxFutureReserve time.Duration) Reservation {
	lim.mu.Lock()

	if lim.limit == Inf {
		lim.mu.Unlock()
		return Reservation{
			ok:        true,
			lim:       lim,
			tokens:    n,
			timeToAct: now,
		}
	}
	// 计算lim.last到now这段时间，tokens的总数是多少
	// tokens := lim.tokens + delta
	// last == lim.last或者 时间回退时等于now
	now, last, tokens := lim.advance(now)

	// Calculate the remaining number of tokens resulting from the request.
	// 扣除本次分配的n个token后，剩余的tokens数目,即未来的 lim.tokens
	tokens -= float64(n)

	// Calculate the wait duration
	var waitDuration time.Duration
	//  如果tokens < 0, 相当于预支了 未来一段时间 新生成的token，因此，需要指定 等待时间。
	//  上游代码 必须等待 waitDuration后，才能满足本次n个token的分配
	// 当tokens >=0 时， waitDuration=0,则后面的r.timeToAct = now+0
	if tokens < 0 {
		waitDuration = lim.limit.durationFromTokens(-tokens)
	}

	// Decide result
	// 决定最终 申请结果
	//  ok = true，才代表 申请token成功
	ok := n <= lim.burst && waitDuration <= maxFutureReserve

	// Prepare reservation
	r := Reservation{
		ok:    ok,
		lim:   lim,
		limit: lim.limit,
	}
	if ok {
		r.tokens = n
		r.timeToAct = now.Add(waitDuration)
	}

	// Update state
	if ok {
		lim.last = now
		lim.tokens = tokens
		lim.lastEvent = r.timeToAct
	} else {
		lim.last = last
	}

	lim.mu.Unlock()
	return r
}

// advance calculates and returns an updated state for lim resulting from the passage of time.
// lim is not changed.
// advance requires that lim.mu is held.
// 计算 从lim.last到newNow期间，新生成的token增量delta，然后算出令牌桶预期的tokens大小
func (lim *Limiter) advance(now time.Time) (newNow time.Time, newLast time.Time, newTokens float64) {
	last := lim.last
	// 说明了时间回退，则把last标记为now，elapsed =0, delta=0, tokens不变
	// 即时间回退时， tokens数目不变  参见 TestLimiterJumpBackwards
	// 时间回退时，即使申请失败， 也会改变lim.last
	if now.Before(last) {
		last = now
	}

	// Avoid making delta overflow below when last is very old.
	// 计算 从lim.tokens 经过多少时长才能增长到lim.burst
	maxElapsed := lim.limit.durationFromTokens(float64(lim.burst) - lim.tokens)
	elapsed := now.Sub(last)
	if elapsed > maxElapsed {
		elapsed = maxElapsed
	}

	// Calculate the new number of tokens, due to time that passed.
	delta := lim.limit.tokensFromDuration(elapsed)
	tokens := lim.tokens + delta
	if burst := float64(lim.burst); tokens > burst {
		tokens = burst
	}

	return now, last, tokens
}
