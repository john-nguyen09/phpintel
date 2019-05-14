<?php
// This file is part of Moodle - http://moodle.org/
//
// Moodle is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Moodle is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Moodle.  If not, see <http://www.gnu.org/licenses/>.

/**
 * Contains classes, functions and constants used during the tracking
 * of activity completion for users.
 *
 * Completion top-level options (admin setting enablecompletion)
 *
 * @package core_completion
 * @category completion
 * @copyright 1999 onwards Martin Dougiamas   {@link http://moodle.com}
 * @license http://www.gnu.org/copyleft/gpl.html GNU GPL v3 or later
 */

defined('MOODLE_INTERNAL') || die();

/**
 * Include the required completion libraries
 */
require_once $CFG->dirroot.'/completion/completion_aggregation.php';
require_once $CFG->dirroot.'/completion/criteria/completion_criteria.php';
require_once $CFG->dirroot.'/completion/completion_completion.php';
require_once $CFG->dirroot.'/completion/completion_criteria_completion.php';


/**
 * The completion system is enabled in this site/course
 */
define('COMPLETION_ENABLED', 1);
/**
 * The completion system is not enabled in this site/course
 */
define('COMPLETION_DISABLED', 0);

/**
 * Completion tracking is disabled for this activity
 * This is a completion tracking option per-activity  (course_modules/completion)
 */
define('COMPLETION_TRACKING_NONE', 0);

/**
 * Manual completion tracking (user ticks box) is enabled for this activity
 * This is a completion tracking option per-activity  (course_modules/completion)
 */
define('COMPLETION_TRACKING_MANUAL', 1);
/**
 * Automatic completion tracking (system ticks box) is enabled for this activity
 * This is a completion tracking option per-activity  (course_modules/completion)
 */
define('COMPLETION_TRACKING_AUTOMATIC', 2);

/**
 * The user has not completed this activity.
 * This is a completion state value (course_modules_completion/completionstate)
 */
define('COMPLETION_INCOMPLETE', 0);
/**
 * The user has completed this activity. It is not specified whether they have
 * passed or failed it.
 * This is a completion state value (course_modules_completion/completionstate)
 */
define('COMPLETION_COMPLETE', 1);
/**
 * The user has completed this activity with a grade above the pass mark.
 * This is a completion state value (course_modules_completion/completionstate)
 */
define('COMPLETION_COMPLETE_PASS', 2);
/**
 * The user has completed this activity but their grade is less than the pass mark
 * This is a completion state value (course_modules_completion/completionstate)
 */
define('COMPLETION_COMPLETE_FAIL', 3);

/**
 * The effect of this change to completion status is unknown.
 * A completion effect changes (used only in update_state)
 */
define('COMPLETION_UNKNOWN', -1);
/**
 * The user's grade has changed, so their new state might be
 * COMPLETION_COMPLETE_PASS or COMPLETION_COMPLETE_FAIL.
 * A completion effect changes (used only in update_state)
 */
define('COMPLETION_GRADECHANGE', -2);

/**
 * User must view this activity.
 * Whether view is required to create an activity (course_modules/completionview)
 */
define('COMPLETION_VIEW_REQUIRED', 1);
/**
 * User does not need to view this activity
 * Whether view is required to create an activity (course_modules/completionview)
 */
define('COMPLETION_VIEW_NOT_REQUIRED', 0);

/**
 * User has viewed this activity.
 * Completion viewed state (course_modules_completion/viewed)
 */
define('COMPLETION_VIEWED', 1);
/**
 * User has not viewed this activity.
 * Completion viewed state (course_modules_completion/viewed)
 */
define('COMPLETION_NOT_VIEWED', 0);

/**
 * Completion details should be ORed together and you should return false if
 * none apply.
 */
define('COMPLETION_OR', false);
/**
 * Completion details should be ANDed together and you should return true if
 * none apply
 */
define('COMPLETION_AND', true);

/**
 * Course completion criteria aggregation method.
 */
define('COMPLETION_AGGREGATION_ALL', 1);
/**
 * Course completion criteria aggregation method.
 */
define('COMPLETION_AGGREGATION_ANY', 2);
