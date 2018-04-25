<?php
declare(strict_types=1);
namespace PhpIntel;

use PhpIntel\Protocol\Location;

abstract class Symbol
{
    /**
     *
     * @var Location
     */
    public $location;

    /**
     *
     * @var string
     */
    public $name;

    /**
     * Fully qualified name
     *
     * @var string[]
     */
    public $types;

    public function __construct(Location $location, string $name, array $types)
    {
        $this->location = $location;
        $this->name = $name;
        $this->types = $types;
    }
}