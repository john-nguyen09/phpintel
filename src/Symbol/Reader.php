<?php
namespace PhpIntel\Symbol;

use Microsoft\PhpParser\Node;
use Microsoft\PhpParser\Node\DelimitedList;
use Microsoft\PhpParser\Node\Expression;
use Microsoft\PhpParser\Node\Statement;
use Microsoft\PhpParser\ResolvedName;
use Microsoft\PhpParser\Token;
use Microsoft\PhpParser\TokenKind;
use PhpIntel\Entity;
use PhpIntel\PhpDocument;
use PhpIntel\NodeVisitor;
use PhpIntel\ParserHelper;
use PhpIntel\Protocol\Location;
use PhpIntel\Protocol\Range;
use PhpIntel\Protocol\Position;
use PhpIntel\Symbol;
use PhpIntel\Symbol\Type\Resolver;
use PhpIntel\Symbol\Type\FunctionResolver;

class Reader extends NodeVisitor
{
    /**
     * @var string[]
     */
    private $scopeStack = [];

    /**
     * @var int
     */
    private $scopeStackIndex = 0;

    public function before(PhpDocument $doc, $node)
    {
        if ($node instanceof Statement\NamespaceDefinition) {
            return $this->readNamespace($doc, $node);
        } else if ($node instanceof Statement\NamespaceUseDeclaration) {
            return $this->readNamespaceUse($doc, $node);
        } else if ($node instanceof Node\QualifiedName) {
            $name = $node->getText();
            $lcName = strtolower($name);
            
            if ($lcName === '\define' || $lcName === 'define') {
                return $this->readDefine($doc, $node);
            }
        } else if ($node instanceof Statement\ConstDeclaration) {
            return $this->readConstant($doc, $node);
        } else if ($node instanceof Statement\FunctionDeclaration) {
            return $this->readFunction($doc, $node);
        } else if ($node instanceof Statement\ClassDeclaration) {
            return $this->readClass($doc, $node);
        } else if ($node instanceof Statement\InterfaceDeclaration) {
            return $this->readInterface($doc, $node);
        } else if ($node instanceof Statement\TraitDeclaration) {
            return $this->readTrait($doc, $node);
        } else if ($node instanceof Node\PropertyDeclaration) {
            return $this->readProperty($doc, $node);
        } else if ($node instanceof Node\MethodDeclaration) {
            return $this->readMethod($doc, $node);
        } else if ($node instanceof Node\ClassConstDeclaration) {
            return $this->readClassConstant($doc, $node);
        }
    }

    private function pushScope(string $scope) {
        $this->scopeStack[$this->scopeStackIndex] = $scope;
    }

    private function popScope() : string {
        $scope = $this->getScope();
        $this->scopeStackIndex--;

        return $scope;
    }

    private function getScope() : string {
        if ($this->scopeStackIndex < 0) {
            return '';
        }

        return $this->scopeStack[$this->scopeStackIndex];
    }

    protected function getParameters(
        PhpDocument $doc,
        Node\DelimitedList\ParameterDeclarationList $parametersNode = null
    ) {
        if ($parametersNode === null) {
            return [];
        }

        $parameters = [];

        foreach ($parametersNode->children as $parameterNode) {
            if ($parameterNode instanceof Token) {
                continue;
            }

            /**
             * @var Node\Parameter $parameterNode
             */

            $types = [];
            $typeNode = $parameterNode->typeDeclaration;
            $value = null;

            if ($typeNode !== null) {
                if ($typeNode instanceof Token) {
                    $types[] = $parameterNode->typeDeclaration->getText($doc->text);
                } else if ($typeNode instanceof Node\QualifiedName) {
                    /**
                     * @var Node\QualifiedName $typeNode
                     */
                    $types[] = (string) ResolvedName::buildName(
                        $typeNode->getNameParts(), $doc->text
                    );
                }
            }

            if ($parameterNode->default !== null) {
                $value = $parameterNode->default->getText();

                $types[] = Resolver::resolveExpressionToType($parameterNode->default);
            }

            $parameters[] = new Entity\Parameter(
                $types,
                $parameterNode->variableName->getText($doc->text),
                $value
            );
        }

        return $parameters;
    }

    protected function getNames(DelimitedList $nameList) {
        $names = [];

        foreach ($nameList->getElements() as $name) {
            if ($name instanceof Node) {
                $names[] = $name->getText();
            }
        }

        return $names;
    }

    protected function readNamespace(PhpDocument $doc, Statement\NamespaceDefinition $node)
    {
        $doc->setNamespace($node->name->getText());

        return false;
    }

    protected function readNamespaceUse(
        PhpDocument $doc, Statement\NamespaceUseDeclaration $node
    ) {
        if (!isset($node->useClauses)) {
            return false;
        }

            // TODO fix getValues
        foreach ($node->useClauses->getValues()  as $useClause) {
            /**
             * @var Node\NamespaceUseClause $useClause
             */

            $namespaceNamePartsPrefix =
                $useClause->namespaceName !== null ? $useClause->namespaceName->nameParts : [];

            if ($useClause->groupClauses !== null && $useClause instanceof Node\NamespaceUseClause) {
                // use A\B\C\{D\E};             namespace import: ["E" => [A,B,C,D,E]]
                // use A\B\C\{D\E as F};        namespace import: ["F" => [A,B,C,D,E]]
                // use function A\B\C\{A, B}    function import: ["A" => [A,B,C,A], "B" => [A,B,C]]
                // use function A\B\C\{const A} const import: ["A" => [A,B,C,A]]
                foreach ($useClause->groupClauses->children as $groupClause) {
                    if (!($groupClause instanceof Node\NamespaceUseGroupClause)) {
                        continue;
                    }
                    $namespaceNameParts = \array_merge(
                        $namespaceNamePartsPrefix,
                        $groupClause->namespaceName->nameParts
                    );
                    $functionOrConst = $groupClause->functionOrConst ?? $node->functionOrConst;
                    $alias = $groupClause->namespaceAliasingClause === null
                        ? $groupClause->namespaceName->getLastNamePart()->getText($doc->text)
                        : $groupClause->namespaceAliasingClause->name->getText($doc->text);

                    $doc->addToImportTable(
                        $alias,
                        $functionOrConst,
                        $namespaceNameParts
                    );
                }
            } else {
                // use A\B\C;               namespace import: ["C" => [A,B,C]]
                // use A\B\C as D;          namespace import: ["D" => [A,B,C]]
                // use function A\B\C as D  function import: ["D" => [A,B,C]]
                // use A\B, C\D;            namespace import: ["B" => [A,B], "D" => [C,D]]
                $alias = $useClause->namespaceAliasingClause === null
                    ? $useClause->namespaceName->getLastNamePart()->getText($doc->text)
                    : $useClause->namespaceAliasingClause->name->getText($doc->text);
                $functionOrConst = $node->functionOrConst;
                $namespaceNameParts = $namespaceNamePartsPrefix;

                $doc->addToImportTable(
                    $alias,
                    $functionOrConst,
                    $namespaceNameParts
                );
            }
        }

        return false;
    }

    protected function readDefine(PhpDocument $doc, $node)
    {
        /**
         * @var Expression\CallExpression $callExpr
         */
        $callExpr = $node->parent;

        $name = $callExpr->argumentExpressionList->children[0]->expression
            ->getStringContentsText();
        $valueNode = $callExpr->argumentExpressionList->children[2]->expression;

        $doc->addSymbol(new DefineConstantSymbol(
            Location::fromNode($doc, $node),
            $name,
            Resolver::resolveExpressionToType($valueNode),
            $valueNode->getText()
        ));

        return false;
    }

    protected function readConstant(PhpDocument $doc, Statement\ConstDeclaration $node)
    {
        foreach ($node->constElements->getElements() as $constElement) {
            if ($constElement instanceof Node\ConstElement) {
                $doc->addSymbol(new ConstantSymbol(
                    Location::fromNode($doc, $constElement),
                    $constElement->name->getText($doc->text),
                    Resolver::resolveExpressionToType($constElement->assignment),
                    $constElement->assignment->getText()
                ));
            }
        }

        return false;
    }

    protected function readFunction(PhpDocument $doc, Statement\FunctionDeclaration $node)
    {
        if ($node->name === null) {
            return;
        }

        $name = $node->name->getText($doc->text);
        $parameters = $this->getParameters($doc, $node->parameters);

        $doc->addSymbol(new FunctionSymbol(
            Location::fromNode($doc, $node),
            $name,
            FunctionResolver::resolveFunctionType($node),
            $parameters
        ));

        // Return false to stop children traversing
        return false;
    }

    protected function readClass(PhpDocument $doc, Statement\ClassDeclaration $node)
    {
        $abstractOrFinal = $node->abstractOrFinalModifier;
        $traversedChildren = [
            'abstractOrFinalModifier'
        ];
        $modifier = Symbol\Modifier::NONE;
        $parent = null;
        $interfaces = [];

        if ($abstractOrFinal !== null) {
            if ($abstractOrFinal->kind == TokenKind::AbstractKeyword) {
                $modifier |= Symbol\Modifier::ABSTRACT;
            } else if ($abstractOrFinal->kind == TokenKind::FinalKeyword) {
                $modifier |= Symbol\Modifier::FINAL;
            }
        }

        if ($node->classBaseClause !== null && $node->classBaseClause->baseClass !== null) {
            $parent = $node->classBaseClause->baseClass->getText();
        }
        $traversedChildren[] = 'classBaseClause';
        if (
            $node->classInterfaceClause !== null &&
            $node->classInterfaceClause->interfaceNameList !== null
        ) {
            $interfaces = $this->getNames($node->classInterfaceClause->interfaceNameList);
        }
        $traversedChildren[] = 'classInterfaceClause';

        $name = $node->name->getText($doc->text);
        $traversedChildren[] = 'name';

        $doc->addSymbol(new ClassSymbol(
            Location::fromNode($doc, $node),
            $name,
            $modifier,
            $parent,
            $interfaces
        ));

        $this->pushScope($name);

        // Do not traverse these children as they are handled in this function
        return $traversedChildren;
    }

    protected function readInterface(PhpDocument $doc, Statement\InterfaceDeclaration $node)
    {
        $traversedChildren = [];
        $parents = [];

        $traversedChildren[] = 'interfaceKeyword';

        if (
            $node->interfaceBaseClause !== null &&
            $node->interfaceBaseClause->interfaceNameList !== null
        ) {
            $parents = $this->getNames($node->interfaceBaseClause->interfaceNameList);
        }
        $traversedChildren[] = 'interfaceBaseClause';

        $name = $node->name->getText($doc->text);

        $doc->addSymbol(new InterfaceSymbol(
            Location::fromNode($doc, $node),
            $name,
            $parents
        ));
        $traversedChildren[] = 'name';

        $this->pushScope($name);

        return $traversedChildren;
    }

    protected function readTrait(PhpDocument $doc, Statement\TraitDeclaration $node)
    {
        $name = $node->name->getText($doc->text);

        $doc->addSymbol(new TraitSymbol(
            Location::fromNode($doc, $node),
            $name
        ));

        $this->pushScope($name);
    }

    protected function readProperty(PhpDocument $doc, Node\PropertyDeclaration $node)
    {
        $modifier = Modifier::NONE;

        foreach ($node->modifiers as $nodeModifier) {
            if ($nodeModifier->kind === TokenKind::PublicKeyword) {
                $modifier |= Modifier::PUBLIC;
            } else if ($nodeModifier->kind === TokenKind::PrivateKeyword) {
                $modifier |= Modifier::PRIVATE;
            } else if ($nodeModifier->kind === TokenKind::StaticKeyword) {
                $modifier |= Modifier::STATIC;
            } else if ($nodeModifier->kind === TokenKind::FinalKeyword) {
                $modifier |= Modifier::FINAL;
            }
        }

        foreach ($node->propertyElements->getElements() as $element) {
            if ($element instanceof Expression\Variable) {
                $propertyName = $element->getName();

                if ($propertyName !== null) {
                    $doc->addSymbol(new PropertySymbol(
                        Location::fromNode($doc, $element),
                        $propertyName,
                        [], // TODO: resolve type based on documentation
                        $modifier,
                        $this->getScope()
                    ));
                }
            }
        }

        return false;
    }

    protected function readMethod(PhpDocument $doc, Node\MethodDeclaration $node)
    {
        if ($node->name === null) {
            return;
        }

        $name = $node->name->getText($doc->text);
        $parameters = $this->getParameters($doc, $node->parameters);
        $parametersNode = $node->parameters;
        $modifier = Modifier::NONE;

        foreach ($node->modifiers as $nodeModifier) {
            if ($nodeModifier->kind === TokenKind::PublicKeyword) {
                $modifier |= Modifier::PUBLIC;
            } else if ($nodeModifier->kind === TokenKind::PrivateKeyword) {
                $modifier |= Modifier::PRIVATE;
            } else if ($nodeModifier->kind === TokenKind::StaticKeyword) {
                $modifier |= Modifier::STATIC;
            } else if ($nodeModifier->kind === TokenKind::FinalKeyword) {
                $modifier |= Modifier::FINAL;
            }
        }

        $doc->addSymbol(new MethodSymbol(
            Location::fromNode($doc, $node),
            $name,
            FunctionResolver::resolveFunctionType($node),
            $parameters,
            $modifier,
            $this->getScope()
        ));

        // Return false to stop children traversing
        return false;
    }

    protected function readClassConstant(PhpDocument $doc, Node\ClassConstDeclaration $node)
    {
        $modifier = Modifier::NONE;

        foreach ($node->modifiers as $nodeModifier) {
            if ($nodeModifier->kind === TokenKind::PublicKeyword) {
                $modifier |= Modifier::PUBLIC;
            } else if ($nodeModifier->kind === TokenKind::PrivateKeyword) {
                $modifier |= Modifier::PRIVATE;
            } else if ($nodeModifier->kind === TokenKind::ProtectedKeyword) {
                $modifier |= Modifier::PROTECTED;
            }
        }

        foreach ($node->constElements->getElements() as $constElement) {
            if ($constElement instanceof Node\ConstElement) {
                $doc->addSymbol(new ClassConstantSymbol(
                    Location::fromNode($doc, $constElement),
                    $constElement->name->getText($doc->text),
                    Resolver::resolveExpressionToType($constElement->assignment),
                    $constElement->assignment->getText(),
                    $modifier,
                    $this->getScope()
                ));
            }
        }

        return false;
    }
}