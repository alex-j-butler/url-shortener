# config valid for current version and patch releases of Capistrano
lock "~> 3.10.1"

set :application, "url-shortener"
set :repo_url, "git@github.com:alex-j-butler/url-shortener.git"

# Default branch is :master
# ask :branch, `git rev-parse --abbrev-ref HEAD`.chomp

# Default deploy_to directory is /var/www/my_app_name
set :deploy_to, "/home/alex/url-shortener"

# Default value for :format is :airbrussh.
# set :format, :airbrussh

# You can configure the Airbrussh format using :format_options.
# These are the defaults.
# set :format_options, command_output: true, log_file: "log/capistrano.log", color: :auto, truncate: :auto

# Default value for :pty is false
# set :pty, true

# Default value for :linked_files is []
append :linked_files, "config.yml"

# Default value for linked_dirs is []
# append :linked_dirs, "log", "tmp/pids", "tmp/cache", "tmp/sockets", "public/system"

# Default value for default_env is {}
# set :default_env, { path: "/opt/ruby/bin:$PATH" }

# Default value for local_user is ENV['USER']
# set :local_user, -> { `git config user.name`.chomp }

# Default value for keep_releases is 5
# set :keep_releases, 5

# Uncomment the following to require manually verifying the host key before first deploy.
# set :ssh_options, verify_host_key: :secure

namespace :deploy do
	desc 'Build'
	task :build do
		run_locally do
			execute 'GOOS=linux GOARCH=amd64 go build -o url-shortener' # Build to bin directory using Go
		end
	end
	after :check, :build

	desc 'Copy to server'
	task :copy do
		archive_name = "archive.tar.gz"
		include_dir = fetch(:include_dir) || "*"

		run_locally do
			files = ['url-shortener'].join(' ')

			execute "tar -cvzf #{archive_name} #{files}"
		end

		on roles(:all) do
			set_release_path
			execute :mkdir, "-p", release_path
			
			tmp_file = capture("mktemp")
			upload!(archive_name, tmp_file)

			File.delete archive_name if File.exists? archive_name

			execute :tar, "-xzf", tmp_file, "-C", release_path
			execute :rm, tmp_file
		end
	end
	after :build, :copy
end
