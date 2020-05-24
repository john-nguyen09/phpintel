<?php

deprecatedFunction(DEPRECATED_CONST);

$deprecatedThing = new DeprecatedClass(EVEN_THOUGH_THIS_IS_DEFINED_BUT_IT_IS_DEPRECATED);
$deprecatedThing->deprecatedProp1 = DeprecatedClass::DEPRECATED_CLASS_CONST;
$deprecatedThing->deprecatedMethod();
DeprecatedClass::deprecatedStaticMethod();
DeprecatedClass::$deprecatedStaticProp;

class InheritingDeprecatedClass extends DeprecatedClass implements DeprecatedInterface
{
    private function method1(DeprecatedClass $instance)
    {
        
    }
}
