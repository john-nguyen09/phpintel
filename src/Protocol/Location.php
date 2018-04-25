<?php

namespace PhpIntel\Protocol;

use Microsoft\PhpParser;
use Microsoft\PhpParser\Node;
use PhpIntel\PhpDocument;

/**
 * Represents a location inside a resource, such as a line inside a text file.
 */
class Location
{
    /**
     * @var string
     */
    public $uri;

    /**
     * @var Range
     */
    public $range;

    /**
     * Returns the location of the node
     *
     * @param Node $node
     * @return self
     */
    public static function fromNode(PhpDocument $doc, Node $node)
    {
        return new self($doc->uri, Range::fromNode($doc, $node));
    }

    public function __construct(string $uri = null, Range $range = null)
    {
        $this->uri = $uri;
        $this->range = $range;
    }
}