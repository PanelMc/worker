/*
 * Default properties when creating a new server.
 * 
 * These can be overriten by Server Preset or by the
 * options passed when creating the server.
 */
server {
    bind {
        host_dir = "/servers/data/%s/"
        volume   = "/data"
    }

    bind {
        host_dir = "/servers/data/%s-plugins/"
        volume   = "/plugins"
    }
}

presets_folder = "./presets/"

// Permission used when creating a new file. e.g. configuration files
file_permissions = 644
// Permission used when creating a new folder
folder_permissions = 744
