<?php

class ClassWithMethod
{
    public static $staticVariable = false;

    public $property1 = false;
    public $property2;

    /**
     * Defines protected properties
     * 
     * @var boolean $protected1 Description of boolean
     * @var string $protected2 Description of string
     * @var int $protected3 Description of int
     */
    protected $protected1, $protected2, $protected3;

    protected $protectedProperty1;

    /**
     * A private proterty which is only used by this class
     * 
     * @var ClassWithMethod This is a description
     */
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

    public static function staticMethod($param1, $param2) {
        return true;
    }
}