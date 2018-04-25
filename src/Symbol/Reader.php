<?php
namespace PhpIntel\Symbol;

use Microsoft\PhpParser\Node;
use Microsoft\PhpParser\Node\DelimitedList;
use Microsoft\PhpParser\Node\Expression;
use Microsoft\PhpParser\Node\Statement;
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
        if ($node instanceof Node\QualifiedName) {
            $name = $node->getText();
            $lcName = strtolower($name);
            
            if ($lcName === '\define' || $lcName === 'define') {
                return $this->readDefine($doc, $node);
            }
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
        }
    }

    private function addSymbol(PhpDocument $doc, Symbol $symbol)
    {
        // TODO: Index the symbol
        $doc->addSymbol($symbol);
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

            $type = null;
            $typeNode = $parameterNode->typeDeclaration;
            $value = null;

            if ($typeNode !== null) {
                if ($typeNode instanceof Token) {
                    $type = $parameterNode->typeDeclaration->getText($doc->text);
                } else if ($typeNode instanceof Node\QualifiedName) {
                    /**
                     * @var Node\QualifiedName $typeNode
                     */
                    $type = '';
                    $stringParts = [];
                    foreach ($typeNode->getNameParts() as $part) {
                        if ($part->kind === TokenKind::Name) {
                            $stringParts[] = $part->getText($doc->text);
                        }
                    }

                    $type = implode('\\', $stringParts);
                }
            }

            if ($parameterNode->default !== null) {
                $value = $parameterNode->default->getText();

                if ($type !== null) {
                    $type = Resolver::resolveExpressionToType($parameterNode->default);
                }
            }

            $parameters[] = new Entity\Parameter(
                $type,
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

    protected function readDefine(PhpDocument $doc, $node)
    {
        /**
         * @var Expression\CallExpression $callExpr
         */
        $callExpr = $node->parent;

        $name = $callExpr->argumentExpressionList->children[0]->expression
            ->getStringContentsText();
        $valueNode = $callExpr->argumentExpressionList->children[2]->expression;

        $this->addSymbol($doc ,new DefineConstantSymbol(
            Location::fromNode($doc, $node),
            $name,
            Resolver::resolveExpressionToType($valueNode),
            $valueNode->getText()
        ));

        return false;
    }

    protected function readFunction(PhpDocument $doc, Statement\FunctionDeclaration $node)
    {
        if ($node->name === null) {
            return;
        }

        $name = $node->name->getText($doc->text);
        $parameters = $this->getParameters($doc, $node->parameters);

        $this->addSymbol($doc, new FunctionSymbol(
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

        $this->addSymbol($doc, new ClassSymbol(
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

        $this->addSymbol($doc, new InterfaceSymbol(
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

        $this->addSymbol($doc, new TraitSymbol(
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
                    $this->addSymbol($doc, new PropertySymbol(
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

        $this->addSymbol($doc, new MethodSymbol(
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
}