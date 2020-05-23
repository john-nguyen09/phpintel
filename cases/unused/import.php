<?php
namespace App1;

use TestClass1;
use TestClass2;
use Namespace1\{
    Test1, Test2, Test3, Test4, Test5
};
use function Namespace2\{
    function1, function2
};
use const Namespace3\{
    CONST1, CONST2
};
use TestClass3;

/** @var Test4 $var1 */
$var1 = TestClass1::doSomething();
$test2 = new Test2();

class ExtendedTest3 extends Test3
{

}

function1(CONST2);
if (Test5\innerFunction()) {
    echo 'Yay!!!' . TestClass3\SOME_CONST;
}
