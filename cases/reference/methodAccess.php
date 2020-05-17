<?php

$var1 = new TestMethodClass();
$var1->testMethod3();

class TestMethodClass2
{
    public function method1()
    {

    }

    public function method2()
    {
        $this->method1();
    }
}
