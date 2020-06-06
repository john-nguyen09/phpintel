<?php

class BaseClass
{
    protected static $baseProtectedStatic = "baseProtectedStatic\n";
    public static $basePublicStatic = "basePublicStatic\n";
    protected $baseProtected = "baseProtected\n";
    protected $baseProtected2 = "baseProtected2\n";
    private $basePrivate = "basePrivate\n";
    public $basePublic = "basePublic\n";
    public $basePublic2 = "basePublic2\n";
}

class ExtendedClass extends BaseClass
{
    protected $baseProtected = "Overriden baseProtected\n";
    public static $basePublicStatic = "Overriden basePublicStatic\n";
    public $basePublic = "Overriden basePublic\n";

    public function main()
    {
        print parent::$baseProtectedStatic;
        print parent::$basePublicStatic;
        print $this->baseProtected;
        print $this->baseProtected2;
        print static::$basePublicStatic;
    }
}

$instance = new ExtendedClass();
print $instance::$basePublicStatic;
print $instance->basePublic;
print $instance->basePublic2;
$base = new BaseClass();
$base->basePublic;
$instance->main();
