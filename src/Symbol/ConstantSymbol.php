<?php
declare(strict_types=1);
namespace PhpIntel\Symbol;

use PhpIntel\PhpDocument;
use PhpIntel\Symbol;
use PhpIntel\Protocol\Location;

class ConstantSymbol extends Symbol
{

    /**
     * Fully qualified name
     *
     * @var string
     */
    public $type;

    /**
     * @var string
     */
    public $value;

    public function __construct(
        Location $location,
        string $name,
        string $type,
        string $value
    ) {
        parent::__construct($location, $name);

        $this->type = $type;
        $this->value = $value;
    }

    public function resolveToFqn(PhpDocument $doc)
    {
        $this->appendNamespaceToName($doc);

        $this->type = $this->aliasToFqn($doc, $this->type);
    }
}