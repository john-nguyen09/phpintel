<?php

class TestConstClass
{
    const CONST1 = 'CONST1';

    public function __construct()
    {
        static::CONST1;
    }
}

TestConstClass::CONST1;
