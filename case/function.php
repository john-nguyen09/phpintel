<?php

function TestFunc1($param1, $param2, $param3 = true, $param4 = '') {

}

function TestFunc2($param1 = null, int $param2) {
    return 1;
}

function TestFunc3($param1, $param2, $param3, bool $param4) {
    if ($param1) {
        return true;
    }

    if ($param2) {
        return 3.14;
    }

    if ($param3) {
        return 15;
    }

    if (!$param4) {
        return null;
    }

    return $param4;
}
