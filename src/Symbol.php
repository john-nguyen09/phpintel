<?php
declare(strict_types=1);
namespace PhpIntel;

use PhpIntel\Protocol\Location;

abstract class Symbol
{
    public const DEFAULT_TYPES = [
        'string',
        'int',
        'bool',
        'float',
        'mixed',
        'null'
    ];

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

    public function __construct(Location $location, string $name)
    {
        $this->location = $location;
        $this->name = $name;
    }

    public abstract function resolveToFqn(PhpDocument $doc);

    protected function appendNamespaceToName(PhpDocument $doc)
    {
        $this->name = $doc->namespace . '\\' . $this->name;
    }

    protected function appendNamespace(PhpDocument $doc, string $name) : string
    {
        return $doc->namespace . '\\' . $name;
    }

    /**
     * Resolve alias to fqn ignoring default types and null alias
     *
     * @param PhpDocument $doc
     * @param string $alias
     * @return string|null
     */
    protected function aliasToFqn(PhpDocument $doc, string $alias = null)
    {
        if ($alias === null) {
            return null;
        }

        // Do not resolve default types
        if (\in_array($alias, self::DEFAULT_TYPES)) {
            return $alias;
        }

        $aliasNameParts = explode('\\', $alias);

        // No nested alias
        // use A\B\C; new C => 'A\B\C'
        // namespace A\B; new C => 'A\B\C'
        if (\count($aliasNameParts) === 1) {
            return isset($doc->namespaceImportTable[$alias]) ?
                $doc->namespaceImportTable[$alias] : $this->appendNamespace($doc, $alias);
        }

        // use A\B\C; new C\D => 'A\B\C\D'
        // namespace A\B; new C\D => 'A\B\C\D'
        $firstNamePart = array_shift($aliasNameParts);

        $baseNamespace = isset($doc->namespaceImportTable[$alias]) ?
            $doc->namespaceImportTable[$alias] : $doc->namespace;
        
        array_unshift($aliasNameParts, $baseNamespace);

        return implode('\\', $aliasNameParts);
    }
}