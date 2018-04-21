<?php
namespace PhpIntel\Symbol;

use Microsoft\PhpParser\Node;
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
use PhpIntel\Symbol\Type\Resolver;
use PhpIntel\Symbol\Type\FunctionResolver;

class Reader extends NodeVisitor
{
    /**
     * @var BaseSymbol[]
     */
    public $symbols = [];

    public function before(PhpDocument $doc, $node)
    {
        if ($node instanceof Node\QualifiedName) {
            $name = $node->getText();
            $lcName = strtolower($name);
            
            if ($lcName === '\define' || $lcName === 'define') {
                $this->readDefine($doc, $node);                

                return false;
            }
        } else if ($node instanceof Statement\FunctionDeclaration) {
            $this->readFunction($doc, $node);
        }
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

        $this->symbols[] = new DefineConstantSymbol(
            Location::fromNode($doc, $node),
            $name,
            Resolver::resolveExpressionToType($valueNode),
            $valueNode->getText()
        );
    }

    protected function readFunction(PhpDocument $doc, Statement\FunctionDeclaration $node)
    {
        if ($node->name === null) {
            return;
        }

        $name = $node->name->getText($doc->text);
        $parameters = [];
        $parametersNode = $node->parameters;

        if ($parametersNode !== null) {
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
                    $type, $parameterNode->variableName->getText($doc->text), $value
                );
            }
        }

        $this->symbols[] = new FunctionSymbol(
            Location::fromNode($doc, $node),
            $name,
            FunctionResolver::resolveFunctionType($node),
            $parameters
        );

        return false;
    }
}