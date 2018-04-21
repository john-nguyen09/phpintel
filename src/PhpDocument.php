<?php
declare(strict_types=1);

namespace PhpIntel;

use Microsoft\PhpParser\Node\SourceFileNode;
use PhpIntel\Protocol\Position;

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

        while ($left < $right) {
            $mid = (int)(($left + $right) / 2);
            if ($offset > $this->linesIndex[$mid]) {
                $left = $mid + 1;
            } else if ($offset < $this->linesIndex[$mid]) {
                $right = $mid - 1;
            } else {
                return $mid + 1;
            }
        }

        return $left + 1;
    }

    public function getPositionByOffset(int $offset) : Position
    {
        $line = $this->getLineByOffset($offset);

        return new Position($line, $offset - $this->linesIndex[$line - 1]);
    }

    public function getNumOfLines()
    {
        return \count($this->linesIndex);
    }
}