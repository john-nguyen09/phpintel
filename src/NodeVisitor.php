<?php
namespace PhpIntel;

abstract class NodeVisitor
{
    public abstract function before(PhpDocument $doc, $node);
}