<?php

class test_method_call {
    private function method1($param1, $param2) {
        $this->method1();
    }

    public static function method2($param1, $param2, boolean $param3) {
        return new test_method_call();
    }
}

test_method_call::method2();
