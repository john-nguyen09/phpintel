<?php

function test_function1(string $stringParam, int $intParam, bool $boolParam, float $floatParam) {
    $test = 1;

    if (is_string($test)) {
        return true;
    }

    switch ($test) {
        case 1:
            return $stringParam;
        case 2:
            return $intParam;
        case 3:
            return $boolParam;
        case 4:
            return $floatParam;
        default:
            return false;
    }

    if (is_object($test)) {
        return new TestClass();
    }

    if (is_numeric($test)) {
        return 20;
    }

    return 'string';
}

/**
 * Test function 2
 * @param array|int $param1 array of something or int
 * @param string[] $param2 array of string
 */
function test_function2($param1, $param2) {

}

$callback = function() {
    
};