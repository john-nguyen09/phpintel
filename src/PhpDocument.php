<?php
declare(strict_types=1);

namespace PhpIntel;

use Microsoft\PhpParser\Node\SourceFileNode;
use PhpIntel\Protocol\Position;
use PhpIntel\Symbol;
use Microsoft\PhpParser\ResolvedName;
use Microsoft\PhpParser\Token;

class PhpDocument
{
    /**
     * Uri of the document
     *
     * @var string
     */
    public $uri;

    /**
     * @var SourceFileNode
     */
    public $ast;

    /**
     * @var string
     */
    public $text;

    /**
     * Symbols defined in this document
     *
     * @var Symbol[]
     */
    public $symbols = [];

    /**
     * Document namespace
     *
     * @var string
     */
    public $namespace = null;

    /**
     * Namespace import table in the document, the format of this is
     * Key:   string - alias
     * Value: string - fully qualified name
     *
     * @var string[]
     */
    public $namespaceImportTable = [];

    /**
     * Function import table
     *
     * @var string[]
     */
    public $functionImportTable = [];

    /**
     * Constant import table
     *
     * @var array
     */
    public $constantImportTable = [];

    /**
     * Index of end of lines
     *
     * @var int[]
     */
    protected $linesIndex;

    public function __construct(string $path)
    {
        $this->uri = Uri\pathToUri($path);
        $text = file_get_contents($path);

        $this->indexLinesEnding($text);
        $this->text = $text;
        $this->ast = App::make('parser')->parseSourceFile($text, $this->uri);
    }

    public function getAST() : SourceFileNode
    {
        return $this->ast;
    }

    private function indexLinesEnding(string $text)
    {
        $length = \mb_strlen($text);
        $this->linesIndex = [];
        $lineIndex = 0;

        $this->linesIndex[] = $lineIndex; // First line starts at the beginning
        // Ignore the last character since if it is a new line character,
        // we still do not have another line
        for ($i = 0; $i < $length - 1; $i++) {
            $ch = $text[$i];

            // There is a new-line character, Windows and Linux
            if (
                $ch == "\n" || // Windows and Linux format
                ($ch == "\r" && $text[$i + 1] != "\n") // macOS format
            ) {
                $this->linesIndex[] = $i + 1;
            }
        }
    }

    public function getLineByOffset(int $offset) : int
    {
        $left = 0;
        $right = \count($this->linesIndex) - 1;

        while ($left <= $right) {
            $mid = (int) floor(($left + $right) / 2);
            if ($offset > $this->linesIndex[$mid]) {
                $left = $mid + 1;
            } else if ($offset < $this->linesIndex[$mid]) {
                $right = $mid - 1;
            } else {
                return $mid;
            }
        }

        return $left - 1;
    }

    public function getPositionByOffset(int $offset) : Position
    {
        $line = $this->getLineByOffset($offset);

        return new Position($line + 1, $offset - $this->linesIndex[$line]);
    }

    public function getNumOfLines()
    {
        return \count($this->linesIndex);
    }

    public function addSymbol(Symbol $symbol)
    {
        $symbol->resolveToFqn($this);
        $this->symbols[] = $symbol;
    }

    public function setNamespace(string $namespace)
    {
        $this->namespace = $namespace;
    }

    public function setNamespaceAlias(string $alias, string $fqn)
    {
        $this->namespaceImportTable[$alias] = $fqn;
    }

    public function setFunctionAlias(string $alias, string $fqn)
    {
        $this->functionImportTable[$alias] = $fqn;
    }

    public function setConstantAlias(string $alias, string $fqn)
    {
        $this->constantImportTable[$alias] = $fqn;
    }

    public function addToImportTable(string $alias, $functionOrConst, $namespaceNameParts)
    {
        if ($alias === null) {
            return;
        }

        $fqn = (string) ResolvedName::buildName($namespaceNameParts, $this->text);

        if ($functionOrConst === null) {
            $this->setNamespaceAlias($alias, $fqn);
        } elseif ($functionOrConst->kind === TokenKind::FunctionKeyword) {
            $this->setFunctionAlias($alias, $fqn);
        } elseif ($functionOrConst->kind === TokenKind::ConstKeyword) {
            $this->setConstantAlias($alias, $fqn);
        }
    }
}