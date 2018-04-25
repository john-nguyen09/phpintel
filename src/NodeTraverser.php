<?php
declare(strict_types=1);
namespace PhpIntel;

use Microsoft\PhpParser\Node\SourceFileNode;
use Microsoft\PhpParser\Node;

class NodeTraverser
{
    /**
     *
     * @var NodeVisitor[]
     */
    protected $visitors;

    /**
     * @var PhpDocument[]
     */
    protected $docs;

    /**
     * Count number of doc to index the $docs array
     *
     * @var int
     */
    protected $docCounter;

    public function __construct()
    {
        $this->visitors = [];
        $this->docs = [];
        $this->docCounter = 0;
    }

    public function addVisitor(NodeVisitor $visitor)
    {
        $this->visitors[] = $visitor;
    }

    public function traverse(PhpDocument $doc)
    {
        $docId = $this->docCounter++;
        $this->docs[$docId] = $doc;
        $this->_traverse($docId, $doc->getAST());
    }

    public static function traverseChildren($node, callable $function)
    {
        $function($node);

        if ($node instanceof Node) {
            foreach ($node::CHILD_NAMES as $name) {
                $childNode = $node->$name;

                if ($childNode === null) {
                    continue;
                }

                if (\is_array($childNode)) {
                    foreach ($childNode as $actualChild) {
                        self::traverseChildren($actualChild, $function);
                    }
                } else {
                    self::traverseChildren($childNode, $function);
                }
            }
        }
    }

    private function _traverse(int $docId, $node)
    {
        $doc = $this->docs[$docId];
        $shouldTraverseChildren = $this->visitBefore($doc, $node);

        if ($shouldTraverseChildren !== false && $node instanceof Node) {
            $traversedChildren = [];

            if (\is_array($shouldTraverseChildren)) {
                $traversedChildren = $shouldTraverseChildren;
            }

            foreach ($node::CHILD_NAMES as $name) {
                if (\in_array($name, $traversedChildren)) {
                    continue;
                }

                $childNode = $node->$name;
    
                if ($childNode === null) {
                    continue;
                }
    
                if (\is_array($childNode)) {
                    foreach ($childNode as $actualChild) {
                        $this->_traverse($docId, $actualChild);
                    }
                } else {
                    $this->_traverse($docId, $childNode);
                }
            }
        }
    }

    protected function visitBefore(PhpDocument $doc, $node)
    {
        $traversedChildren = true;

        foreach ($this->visitors as $visitor) {
            $result = $visitor->before($doc, $node);

            // Only if false is returned implicitly
            if ($result === false) {
                $traversedChildren = false;
            } else if (\is_array($result)) {
                $traversedChildren = $result;
            }
        }

        return $traversedChildren;
    }
}
