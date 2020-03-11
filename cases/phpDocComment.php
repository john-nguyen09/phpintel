<?php

/**
 * @param TestClass1 $param1
 * @param TestClass2 $param2
 * @return MasterTestClass
 */
function f1($param1, $param2)
{
    
}

/**
 * @method MasterTestClass method1()
 * @property TestClass1 $hiddenProp1
 * @property-read TestClass2 $readonlyProp1
 * @property-write TestClass1 $writableProp1
 */
class TestClass1
{
    /**
     * @var TestClass2
     */
    public $prop1;
}

/**
 * @global MasterTestClass $GL
 */
global $GL;
