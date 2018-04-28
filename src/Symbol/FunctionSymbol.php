<?php
declare(strict_types=1);
namespace PhpIntel\Symbol;

use PhpIntel\Entity;
use PhpIntel\Symbol;
use PhpIntel\Protocol\Location;
use PhpIntel\PhpDocument;

class FunctionSymbol extends Symbol
{
    use SymbolWithType;

    /**
     * @var Entity\Parameter[]
     */
    public $params;

    /**
     * Fully qualified name
     *
     * @var string[]
     */
    public $types;

    public function __construct(
        Location $location, string $name, array $types, $params = []
    ) {
        parent::__construct($location, $name);

        $this->types = $types;
        $this->params = $params;
    }

    public function resolveToFqn(PhpDocument $doc)
    {
        $this->appendNamespaceToName($doc);
        $this->resolveTypesToFqn($doc);

        foreach ($this->params as $paramKey => $param) {
            foreach ($param->types as $typeKey => $alias) {
                $this->params[$paramKey]->types[$typeKey] = $this->aliasToFqn($doc, $alias);
            }
        }
    }

    public function addParam(Entity\Param $param)
    {
        $this->params[] = $param;
    }
}