<?php

namespace bscitl;

class sss_user {
	/**
	 * Save post data and attachments
	 * @param array $data post data
	 * @param array $files attachments
	 * @return \bscitl\post Post instance
	 */
	private function save_post_data_and_attachments(array $data, $files = null) {
		/* @var $post \block_socialtimeline\post */
		$post = post::get_instance_by_data((object)$data);
		$post->save();
		// Because of socialtime feed sorting requirement (latest modified post goes to the top),
		// timemodified needs to be set to the same as timecreated when creating post
		$post->set_timemodified($post->get_timecreated());
		$post->save();
		
		if (isset($files)) {
			$post->delete_existing_attachments();
			$post->save_attachments($files);
		}		
		return $post;
	}
}
