<?php

abstract class BaseTestClass {
    abstract function abstractMethod() {

    }
}

interface FirstTest {
    public function test1($arg1, bool $arg2);
}

trait Testable {
    public function runTests() {

    }

    public function setUpTest() {

    }
}

trait DatabaseTest {
    public function setUpTest() {

    }
}

class TestClass1 extends BaseTestClass implements FirstTest {
    use Testable, DatabaseTest {
        DatabaseTest::setUpTest insteadof Testable;
    }
}
