1 tab is used to highlight source/script/config text.

As a root run:
    ./setup_server_upload_user.sh
Then vi/nano/... into /etc/ssh/sshd_config 
and add the following lines:
    Match User uploaduser
        ForceCommand internal-sftp
        ChrootDirectory /uploadroot
Then in the same file find line that start with Subsystem and make the folllowing changes:
    #Subsystem      sftp    /usr/libexec/sftp-server # old line
    Subsystem       sftp    internal-sftp # new line
Then restart the ssh service(it can be named sshd in some distros)
    systemctl restart ssh
At that point we've succesfully made chroot jailed sftp-only user uploaduser.

You should download the /uploadroot/builds_upload/.ssh/id_rsa
file to the machine, from which you intend to upload new builds.
This is now your ssh key for the scp.

Thing to keep in mind: my scripts for upload use scp for transfer
but uploaduser is made to be sftp only. Scp in modern distros use sftp under
the hood, but scp in old distros use seperate protol, which is incompatible
with sftp and thus, in that case, transfer script will fail.

After the first succesful build upload, you either simply 
run /uploadroot/builds_upload/new/redeploy.sh manually or setup a cron job.
    */5 * * * * /uploadroot/builds_upload/new/redeploy.sh
