<?php

class BaseClass
{
    protected static function method1()
    {
        print "method1()\n";
    }
    protected function baseProtectedMethod()
    {
        print "baseProtectedMethod()\n";
    }
    private function basePrivateMethod()
    {
        print "basePrivateMethod()\n";
    }
    public static function baseStaticMethod()
    {
        print "baseStaticMethod()\n";
    }
    public function baseMethod()
    {
        print "baseMethod()\n";
    }
    public function baseMethod1()
    {
        print "baseMethod1()\n";
    }
}

class ExtendedClass extends BaseClass
{
    public function method2()
    {
        $this->method1();
        $base = new BaseClass();
        $base->method1();
        $this->baseProtectedMethod();
        $base->baseProtectedMethod();
        static::baseStaticMethod();
        $this->privateMethod();
        $instance = new ExtendedClass();
        $instance->privateMethod();
        $this::baseStaticMethod();
    }

    protected function baseProtectedMethod()
    {
        print "Overriden baseProtectedMethod()\n";
    }

    private function privateMethod()
    {
        print "privateMethod()\n";
    }

    public static function baseStaticMethod()
    {
        print "Overriden baseStaticMethod()\n";
    }

    public function baseMethod()
    {
        print "Overriden baseMethod()\n";
    }
}

// $instance = new BaseClass();
// $instance::method1();

$instance2 = new ExtendedClass();
$instance2->method2();
ExtendedClass::baseStaticMethod();
BaseClass::baseStaticMethod();
