<?php
declare(strict_types=1);
namespace PhpIntel\Symbol;

use PhpIntel\PhpDocument;
use PhpIntel\Symbol;
use PhpIntel\Protocol\Location;

class ClassSymbol extends Symbol
{
    /**
     * The modifier of the class (abstract, final)
     *
     * @var int
     */
    public $modifier;

    /**
     * Parent class
     *
     * @var string
     */
    public $parent;

    /**
     * Implemented interface
     *
     * @var string[]
     */
    public $interfaces;

    public function __construct(
        Location $location,
        string $name,
        int $modifier,
        $parent = null,
        $interfaces = null
    ) {
        parent::__construct($location, $name);

        $this->modifier = $modifier;
        $this->parent = $parent;
        $this->interfaces = $interfaces;
    }

    public function resolveToFqn(PhpDocument $doc)
    {
        $this->appendNamespaceToName($doc);
        $this->parent = $this->aliasToFqn($doc, $this->parent);
        
        foreach ($this->interfaces as $key => $interface) {
            $this->interfaces[$key] = $this->aliasToFqn($doc, $interface);
        }
    }
}
