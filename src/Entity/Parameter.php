<?php
declare(strict_types=1);
namespace PhpIntel\Entity;

class Parameter
{
    /**
     * @var string
     */
    public $type;

    /**
     * @var string
     */
    public $name;

    /**
     * @var string
     */
    public $value;

    public function __construct($type, string $name, $value)
    {
        $this->type = $type;
        $this->name = $name;
        $this->value = $value;
    }
}