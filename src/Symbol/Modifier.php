<?php
namespace PhpIntel\Symbol;

class Modifier
{
    public const NONE = 0;
    public const PUBLIC = 1 << 0;
    public const PROTECTED = 1 << 1;
    public const PRIVATE = 1 << 2;
    public const FINAL = 1 << 3;
    public const ABSTRACT = 1 << 4;
    public const STATIC = 1 << 5;
    public const MAGIC = 1 << 8;
}