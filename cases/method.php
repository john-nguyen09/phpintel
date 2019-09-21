<?php

class TestMethodClass {
    private function testMethod1() {

    }

    protected function testMethod2($param1 = 0, TestAbstractMethodClass $param2) {

    }

    public function testMethod3() {

    }

    private static function testMethod4() {

    }

    protected static function testMethod5() {

    }

    public static function testMethod6() {

    }

    private final function testMethod7() {

    }

    protected final function testMethod8() {

    }

    public final function testMethod9() {

    }
}

abstract class TestAbstractMethodClass {
    private abstract function testMethod10();
    protected abstract function testMethod11();
    public abstract function testMethod12() : TestMethodClass;
}
