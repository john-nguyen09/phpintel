<?php
declare(strict_types=1);
namespace PhpIntel\Test;

use PhpIntel\Entity\Parameter;
use PhpIntel\Protocol\{
    Location, Range, Position
};
use PhpIntel\PhpDocument;
use PhpIntel\Symbol;
use PhpIntel\NodeTraverser;
use PhpIntel\Symbol\Modifier;

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

    public function testReadingNamespace()
    {
        $traverser = new NodeTraverser();
        $docs = [
            'different_namespace' => $this->getPhpDocument('different_namespace.php'),
            'import_table' => $this->getPhpDocument('import_table.php'),
            'namespace' => $this->getPhpDocument('namespace.php'),
            'nested_import_table' => $this->getPhpDocument('nested_import_table.php')
        ];

        $traverser->addVisitor(new Symbol\Reader());

        foreach ($docs as $doc) {
            $traverser->traverse($doc);
        }

        $this->assertEquals(
            $this->getNestedImportTableFileSymbols(), $docs['nested_import_table']->symbols
        );
    }

    private function getNestedImportTableFileSymbols()
    {
        return [
            new Symbol\ClassSymbol(
                new Location(
                    'file:///Users/nana/Documents/Development/phpintel/test/fixture/nested_import_table.php',
                    new Range(new Position(11, 0), new Position(17, 1))
                ),
                'PhpIntel\\Test\\Fixture\\Nested\\NestedImportTable',
                Modifier::NONE,
                null,
                []
            ),
            new Symbol\MethodSymbol(
                new Location(
                    'file:///Users/nana/Documents/Development/phpintel/test/fixture/nested_import_table.php',
                    new Range(new Position(13, 4), new Position(16, 5))
                ),
                'PhpIntel\\Test\\Fixture\\Nested\\doSomething',
                [],
                [
                    new Parameter(
                        ['PhpIntel\\Test\\Fixture\\Nested\\Namespace1\\DifferentNamespaceClass1'],
                        '$arg1',
                        null
                    )
                ],
                Modifier::PUBLIC,
                'PhpIntel\\Test\\Fixture\\Nested\\NestedImportTable'
            )
        ];
    }
}