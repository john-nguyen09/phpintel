<?php

class BaseClass
{
    const BASE_CONST = 'Hello';
}

class ExtendedClass extends BaseClass
{
    const BASE_CONST = 'Overriden';
    const EXTENDED_CONST = 'world';

    public function doSomething()
    {
        print parent::BASE_CONST . ' ' . static::BASE_CONST . "\n";
        print self::BASE_CONST . "\n";
    }
}

print ExtendedClass::BASE_CONST . ' ' . ExtendedClass::EXTENDED_CONST . "\n";
print BaseClass::BASE_CONST . "\n";
$instance = new ExtendedClass();
$instance->doSomething();

