<?php
declare(strict_types=1);
namespace PhpIntel\Test\Fixture\Symbols;

const USER_DELETED = 'deleted';
const USER_ACTIVE = 'active';
const USER_SUSPENDED = 'suspended';

class User extends Model implements Person, Student
{
    public $username;
    public $password;
    public $firstname;
    public $lastname;

    public const PI = 3.14;

    private const SECONDS_IN_HOUR = 60 * 60;

    public function __construct($username, $password = null, $firstname = null, $lastname = null)
    {
        $this->username = $username;
        $this->password = $password;
        $this->firstname = $firstname;
        $this->lastname = $lastname;
    }

    public function getFullname()
    {
        return $this->firstname . ' ' . $this->lastname;
    }

    public function talk()
    {

    }
}

interface Person
{
    public function talk()
    {

    }
}

interface Student
{
    public function study()
    {

    }
}

trait PigBehaviour
{
    public function oink()
    {
        return 'oink';
    }
}