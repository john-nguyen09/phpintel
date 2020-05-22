<?php

/**
 * @deprecated 1.0.0 This will be removed in the next minor release i.e. 1.1.0
 */
function deprecatedFunction()
{

}

/**
 * @deprecated
 */
const DEPRECATED_CONST = '';

/** @deprecated Of course it is deprecated */
define('EVEN_THOUGH_THIS_IS_DEFINED_BUT_IT_IS_DEPRECATED', 1);

/**
 * @deprecated This class is deprecated the moment it's introduced
 */
class DeprecatedClass
{
    /** @deprecated */
    const DEPRECATED_CLASS_CONST = 'This is deprecated by nature';

    /**
     * @deprecated This is deprecated too
     */
    public $deprecatedProp1;

    /** @deprecated */
    public static $deprecatedStaticProp;

    /**
     * @deprecated
     */
    public function deprecatedMethod()
    {
    }

    /**
     * @deprecated
     */
    public static function deprecatedStaticMethod()
    {
    }
}

/** @deprecated */
interface DeprecatedInterface
{
}
