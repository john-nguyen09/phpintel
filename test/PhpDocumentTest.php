<?php
declare(strict_types=1);
namespace PhpIntel\Test;

use Microsoft\PhpParser;
use PhpIntel\PhpDocument;
use PhpIntel\Protocol\Position;

class PhpDocumentTest extends PhpIntelTestCase
{
    public function testLineFunctions()
    {
        /**
         * @var PhpDocument $doc
         */
        $doc = $this->getPhpDocument('lines.php');

        $this->assertEquals(48, $doc->getNumOfLines());
        $this->assertEquals(8, $doc->getLineByOffset(140) + 1);
        $this->assertEquals(38, $doc->getLineByOffset(558) + 1);
        $this->assertEquals(new Position(38, 22), $doc->getPositionByOffset(558));
    }
}