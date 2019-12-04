<?php

User::create([
    'firstname' => 'Admin',
    'lastname' => 'User',
    'email' => 'admin@example.com',
    'username' => 'admin',
    'password' => bcrypt('a_very_secretive_password'),
]);
