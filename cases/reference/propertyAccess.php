<?php

class TestPropertyClass1
{
    public $prop1;
    public static $prop2;

    private function method1()
    {
        $this->prop1;
        static::$prop2;
    }
}

TestPropertyClass1::$prop2;
$var1 = new TestPropertyClass1();
$var1->prop1;
