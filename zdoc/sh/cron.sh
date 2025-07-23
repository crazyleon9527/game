chmod +x /path/to/your/script.sh
crontab -e

* * * * * /data/cheetah/auto/auto_admin_git.sh >> /data/cheetah/auto/admin.log 2>&1
* * * * * /data/cheetah/auto/auto_front_git.sh >> /data/cheetah/auto/front.log 2>&1
* * * * * (sleep 30; /data/cheetah/auto/auto_front_git.sh) >> /data/cheetah/auto/front.log 2>&1