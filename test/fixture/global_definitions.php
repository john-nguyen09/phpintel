<?php

define('STRING_CONST', 'This is a string');
\define('FLOAT_CONST', 3.14);
define('BOOLEAN_CONST', true);

function global_function()
{

}

function function_with_args(array $args, $arg2, DatabaseInterface $arg3)
{

}

class GlobalClass
{
    public function method1()
    {

    }

    public function method2($arg1, $arg2, $arg3)
    {

    }
}

interface DatabaseInterface
{
    public function getRecord($arg1, $arg2);
    public function getRecords($arg1, $arg2, $arg3);
}

class MysqlDatabase implements DatabaseInterface
{
    public function getRecord($arg1, $arg2)
    {

    }

    public function getRecords($arg1, $arg2, $arg3)
    {

    }
}

