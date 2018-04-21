<?php
declare(strict_types=1);
namespace PhpIntel\Test;

use PHPUnit\Framework\TestCase;
use Microsoft\PhpParser\Parser;
use PhpIntel\App;
use PhpIntel\PhpDocument;

class PhpIntelTestCase extends TestCase
{
    protected function setUp()
    {
        parent::setUp();

        App::init();
    }

    /**
     * get the PhpDocument object
     *
     * @param string $fileName
     * @return PhpDocument
     */
    protected function getPhpDocument(string $fileName)
    {
        $filePath = __DIR__ . '/fixture/' . $fileName;

        return new PhpDocument($filePath);
    }
}