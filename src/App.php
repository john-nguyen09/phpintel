<?php
declare(strict_types=1);

namespace PhpIntel;

use League\Container\Container;
use Microsoft\PhpParser\Parser;

final class App
{
    /**
     * @var Container
     */
    private static $container;

    public static function init()
    {
        if (self::$container === null) {
            self::$container = new Container();
        }

        self::bind();
    }

    public static function make(string $alias)
    {
        return self::$container->get($alias);
    }

    protected static function bind()
    {
        self::$container->share('parser', new Parser);
    }
}