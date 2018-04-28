<?php
declare(strict_types=1);

function string_function() {
    return '';
}

function int_function() {
    return 5;
}

function class_function() {
    return new Class1();
}

function boolean_function() {
    return true;
}

function float_function() {
    return 3.14;
}

function composite_types_function() {
    if ($something) {
        return '';
    } else if ($someInt) {
        return 5;
    } else if ($someBool) {
        return false;
    } else if ($someClass) {
        return new Class1;
    }

    return null;
}

class Class1
{

}