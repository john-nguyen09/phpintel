<?php

class ClassWithMethod
{
    public $property1 = false;
    public $property2;

    protected $protected1, $protected2, $protected3;

    protected $protectedProperty1;

    private $privateProperty1;

    function __construct($optional = null) {

    }

    public function method1() {
        return true;
    }

    public function method2() {
        return [
            'abc1',
            'abc2'
        ];
    }

    public function method3() {
        return 5;
    }

    public function method4() {
        return 3.14;
    }

    protected function protectedMethod1($firstParam) {

    }

    private function privateMethod1($first, $second, $optional1 = null, $optional2 = '') {
        return null;
    }
}