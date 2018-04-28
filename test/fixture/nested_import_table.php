<?php
declare(strict_types=1);
namespace PhpIntel\Test\Fixture\Nested;

use PhpIntel\Test\Fixture;
use PhpIntel\Test\Fixture\{
    Namespace1, ImportTableClass
};
use PhpIntel\Test\Class1 as ClassName;

class NestedImportTable
{
    public function doSomething(Fixture\Namespace1\DifferentNamespaceClass1 $arg1)
    {
        $instance = new Fixture\Namespace1\DifferentNamespaceClass2();
    }
}