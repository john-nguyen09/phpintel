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
        $doc = $this->getPhpDocument('function_type.php');
        $symbolReader = new Symbol\Reader();
        $traverser = new NodeTraverser();
        $traverser->addVisitor($symbolReader);
        $traverser->traverse($doc);

        $expectedFunctionType = [
            'string_function' => ['string'],
            'int_function' => ['int'],
            'boolean_function' => ['bool'],
            'float_function' => ['float'],
            'class_function' => ['Class1'],
            'composite_types_function' => ['string', 'int', 'bool', 'Class1', 'null']
        ];

        $functionSymbols = array_filter($doc->symbols, function($symbol) {
            return $symbol instanceof Symbol\FunctionSymbol;
        });

        $this->assertEquals(\count($functionSymbols), \count($expectedFunctionType));

        foreach ($functionSymbols as $symbol) {
            $actualTypes = $symbol->types;
            $expectedTypes = $expectedFunctionType[$symbol->name];

            $this->assertEquals($expectedTypes, $actualTypes);
        }
    }

    public function testReadingClass()
    {
        $symbolReader = new Symbol\Reader();
        $traverser = new NodeTraverser();

        $traverser->addVisitor($symbolReader);
        $docs = [
            $this->getPhpDocument('Model.php'),
            $this->getPhpDocument('User.php')
        ];

        foreach ($docs as $doc) {
            $traverser->traverse($doc);
        }
    }
}