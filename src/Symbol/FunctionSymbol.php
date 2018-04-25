<?php
declare(strict_types=1);
namespace PhpIntel\Symbol;

use PhpIntel\Entity;
use PhpIntel\Symbol;
use PhpIntel\Protocol\Location;

class FunctionSymbol extends Symbol
{
    /**
     * @var Entity\Parameter[]
     */
    public $params;

    public function __construct(
        Location $location, string $name, array $types, $params = []
    ) {
        parent::__construct($location, $name, $types);

        $this->params = $params;
    }

    public function addParam(Entity\Param $param)
    {
        $this->params[] = $param;
    }
}