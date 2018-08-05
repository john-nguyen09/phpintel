<?php

/**
 * This is a class that has description
 */
class ClassWithDoc
{
    /**
     * Description of $foo
     * @var string
     */
    public $foo;

    /**
     * RGB code for blue sky color
     * @var int
     */
    public $bluesky_color_code;

    /**
     * Get the rainbow color based on some preference
     * @param  int $bluesky_color_code The blue sky color code
     * @param  int $your_prefer_color  Your preference RGB color code
     * @return int[]                     All possible RGB color codes of rainbow (length is 7)
     */
    public function getRainbowColors($bluesky_color_code, $your_prefer_color)
    {

    }

    /**
     * Calculate RGB color code based on red, green and blue
     * @param  int $red   Red color range from 0-255
     * @param  int $green
     * @param  int $blue
     * @return int
     */
    private function calculateRGBColor($red, $green, $blue)
    {

    }
}
