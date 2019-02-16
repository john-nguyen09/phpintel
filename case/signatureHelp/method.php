<?php

class test_method_call {
    private function method1($param1, $param2) {
        $this->method1();
    }
}
