#include "include/foo.h"

void fun1()
{
    printf("f1\n");
}

void fun2(int a)
{
    printf("f2, param=%d\n", a);
}

int fun3(void **b)
{
    printf("f3, param=%p\n", b);
    return 10;
}

