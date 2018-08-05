<?php
namespace NamespaceDef;

/**
 * @param TestClassNamespace1 $param1
 * @param \FullyQualified\Name\Class $param2
 * @param boolean
 * 
 * @return TestClassNamespace2
 */
function function_with_namespace($param1, $param2, bool $param3) {
    $var1 = false;
    $var2 = 'abc';
    $var3 = CONSTANT_IN_NAMESPACE;

    if ($var1 && $var2 && $var3) {
        return $var3;
    }

    if ($var1 && $var2) {
        return $var2;
    }

    if ($var1) {
        return $param1;
    }

    if ($var2) {
        return $param2;
    }

    if ($var3) {
        return $param3;
    }
}