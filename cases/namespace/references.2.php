<?php

namespace TestNamespace1;

use TestClass as RootTestClass;
use function TestNamespace3\{
    testFunction2,
    testFunction3
};

$instance1 = new TestClass();
$instance1->method1000();

$instance2 = new RootTestClass();
$instance2->method1();

testFunction();
\testFunction();

testFunction2()->;
testFunction3();

CONST1;
