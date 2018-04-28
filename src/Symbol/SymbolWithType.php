<?php
declare(strict_types=1);
namespace PhpIntel\Symbol;

use PhpIntel\PhpDocument;
use PhpIntel\Symbol;

trait SymbolWithType
{
    protected function resolveTypesToFqn(PhpDocument $doc)
    {
        if (!isset($this->types) || !($this instanceof Symbol)) {
            return;
        }

        foreach ($this->types as $key => $alias) {
            $this->types[$key] = $this->aliasToFqn($doc, $alias);
        }
    }
}