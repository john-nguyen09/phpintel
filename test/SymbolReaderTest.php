<?php
declare(strict_types=1);
namespace PhpIntel\Test;

use PhpIntel\PhpDocument;
use PhpIntel\Symbol;
use PhpIntel\NodeTraverser;

final class SymbolReaderTest extends PhpIntelTestCase
{
    // public function testReadingSymbols()
    // {
    //     $doc = $this->getPhpDocument('global_definitions.php');
    //     $symbolReader = new Symbol\Reader();
    //     $traverser = new NodeTraverser();

    //     $traverser->addVisitor($symbolReader);

    //     $traverser->traverse($doc);

    //     var_dump($symbolReader->symbols);
    // }

    public function testReadingFunctionType()
    {
        $symbolReader = new Symbol\Reader();
        $traverser = new Nodetraverser();

        $traverser->addVisitor($symbolReader);

        $traverser->traverse($this->getPhpDocument('function_type.php'));

        $expectedFunctionType = [
            'string_function' => ['string'],
            'int_function' => ['int'],
            'boolean_function' => ['bool'],
            'float_function' => ['float'],
            'class_function' => ['Class1'],
            'composite_types_function' => ['string', 'int', 'bool', 'Class1', 'null']
        ];

        foreach ($symbolReader->symbols as $symbol) {
            $actualTypes = $symbol->types;
            $expectedTypes = $expectedFunctionType[$symbol->name];

            $this->assertEquals($expectedTypes, $actualTypes);
        }
    }
}