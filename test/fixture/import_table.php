<?php
declare(strict_types=1);
namespace PhpIntel\Test\Fixture;

use PhpIntel\Test\Fixture\ClassNamedNamespace;
use PhpIntel\Test\Fixture\Namespace1\DifferencenamespaceClass3;

class ImportTableClass
{
    public function method1(ClassNamedNamespace $class) {
        return new DifferencenamespaceClass3();
    }
}