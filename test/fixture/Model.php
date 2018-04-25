<?php
declare(strict_types=1);
namespace PhpIntel\Test\Fixture;

abstract class Model
{
    public function save()
    {
        return $this;
    }

    public static function find($id)
    {
        return new static;
    }
}