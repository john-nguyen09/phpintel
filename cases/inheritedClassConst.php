<?php

class BaseClass
{
    const BASE_CONST = 'Hello';
    protected const BASE_PROTECTED_CONST = 'Base protected const';
    const INHERITED_CONST = 'Inherited const';
}

class ExtendedClass extends BaseClass
{
    const BASE_CONST = 'Overriden';
    private const PRIVATE_CONST = 'A private class const';
    const EXTENDED_CONST = 'world';

    public function doSomething()
    {
        print 'parent && static: ' . parent::BASE_CONST . ' ' . static::BASE_CONST . "\n";
        print 'self: ' . self::BASE_CONST . "\n";
        print 'self private: ' . self::PRIVATE_CONST . "\n";
        $instance = new ExtendedClass();
        print '$instance same class: ' . $instance::PRIVATE_CONST. "\n";
        print '$this static: ' . $this::PRIVATE_CONST . "\n";
        $base = new BaseClass();
        print '$base protected' . $base::BASE_PROTECTED_CONST . "\n";
    }
}

print ExtendedClass::BASE_CONST . ' ' . ExtendedClass::EXTENDED_CONST . "\n";
print ExtendedClass::INHERITED_CONST . "\n";
print BaseClass::BASE_CONST . "\n";
$instance = new ExtendedClass();
$instance->doSomething();

