<?php
declare(strict_types=1);
namespace PhpIntel\Entity;

class Parameter
{
    /**
     * @var string[]
     */
    public $types;

    /**
     * @var string
     */
    public $name;

    /**
     * @var string
     */
    public $value;

    public function __construct(array $types, string $name, $value)
    {
        $this->types = $types;
        $this->name = $name;
        $this->value = $value;
    }
}