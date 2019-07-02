# Security
## Introduction
LXD is a daemon running as root.

Access to that daemon is only possible over a local UNIX socket by default.
Through configuration, it's then possible to expose the same API over
the network on a TLS socket.

**WARNING**: Local access to LXD through the UNIX socket always grants
full access to LXD. This includes the ability to attach any filesystem
paths or devices to any container as well as tweaking all security
features on containers. You should only give such access to someone who
you'd trust with root access to your system.

The remote API uses either TLS client certificates or Candid based
authentication. Canonical RBAC support can be used combined with Candid
based authentication to limit what an API client may do on LXD.

## TLS configuration
Remote communications with the LXD daemon happen using JSON over HTTPS.
The supported protocol must be TLS1.2 or better.

All communications must use perfect forward secrecy and ciphers must be
limited to strong elliptic curve ones (such as ECDHE-RSA or ECDHE-ECDSA).

Any generated key should be at least 4096bit RSA, preferably EC384 and
when using signatures, only SHA-2 signatures should be trusted.

Since we control both client and server, there is no reason to support
any backward compatibility to broken protocol or ciphers.

Both the client and the server will generate a keypair the first time
they're launched. The server will use that for all https connections to
the LXD socket and the client will use its certificate as a client
certificate for any client-server communication.

To cause certificates to be regenerated, simply remove the old ones. On the
next connection a new certificate will be generated.

## Role Based Access Control (RBAC)
LXD supports integrating with the Canonical RBAC service.

This uses Candid based authentication with the RBAC service maintaining
roles to user/group relationships. Roles can be assigned to individual
projects, to all projects or to the entire LXD instance.

The meaning of the roles when applied to a project is as follow:

 - auditor: Read-only access to the project
 - user: Ability to do normal lifecycle actions (start, stop, ...),
   execute commands in the containers, attach to console, manage snapshots, ...
 - operator: All of the above + the ability to create, re-configure and
   delete containers and images
 - admin: All of the above + the ability to reconfigure the project itself

**WARNING**: Of those roles, only `auditor` and `user` are currently
suitable for a user whom you wouldn't trust with root access to the
host.

## Container security
LXD containers can use a pretty wide range of features for security.

By default containers are `unprivileged`, meaning that they operate
inside a user namespace, restricting the abilities of users in the
container to that of regular users on the host with limited privileges
on the devices that the container owns.

If data sharing between containers isn't needed, it is possible to
enable `security.idmap.isolated` which will use non-overlapping uid/gid
maps for each container, preventing potential DoS attacks on other
containers.

LXD can also run `privileged` containers if you so wish, do note that
those aren't root safe and a user with root in such a container will be
able to DoS the host as well as find ways to escape confinement.

More details on container security and the kernel features we use can be found on the
[LXC security page](https://linuxcontainers.org/lxc/security/).

## Adding a remote with TLS client certificate authentication
In the default setup, when the user adds a new server with `lxc remote add`,
the server will be contacted over HTTPS, its certificate downloaded and the
fingerprint will be shown to the user.

The user will then be asked to confirm that this is indeed the server's
fingerprint which they can manually check by connecting to or asking
someone with access to the server to run the info command and compare
the fingerprints.

After that, the user must enter the trust password for that server, if
it matches, the client certificate is added to the server's trust store
and the client can now connect to the server without having to provide
any additional credentials.

This is a workflow that's very similar to that of SSH where an initial
connection to an unknown server triggers a prompt.

## Adding a remote with a TLS client in a PKI based setup
In the PKI setup, a system administrator is managing a central PKI, that
PKI then issues client certificates for all the lxc clients and server
certificates for all the LXD daemons.

Those certificates and keys are manually put in place on the various
machines, replacing the automatically generated ones.

The CA certificate is also added to all machines.

In that mode, any connection to a LXD daemon will be done using the
preseeded CA certificate. If the server certificate isn't signed by the
CA, the connection will simply go through the normal authentication
mechanism.

If the server certificate is valid and signed by the CA, then the
connection continues without prompting the user for the certificate.

After that, the user must enter the trust password for that server, if
it matches, the client certificate is added to the server's trust store
and the client can now connect to the server without having to provide
any additional credentials.

Enabling PKI mode is done by adding a client.ca file in the
client's configuration directory (`~/.config/lxc`) and a server.ca file in
the server's configuration directory (`/var/lib/lxd`). Then a client
certificate must be issued by the CA for the client and a server
certificate for the server. Those must then replace the existing
pre-generated files.

After this is done, restarting the server will have it run in PKI mode.

## Adding a remote with Candid authentication
When LXD is configured with Candid, it will request that clients trying to
authenticating with it get a Discharge token from the authentication server
specified by the `candid.api.url` setting.

The authentication server certificate needs to be trusted by the LXD server.

To add a remote pointing to a LXD configured with Macaroon auth, run `lxc
remote add REMOTE ENDPOINT --auth-type=candid`.  The client will prompt for
the credentials required by the authentication server in order to verify the
user. If the authentication is successful, it will connect to the LXD server
presenting the token received from the authentication server.  The LXD server
verifies the token, thus authenticating the request.  The token is stored as
cookie and is presented by the client at each request to LXD.

## Managing trusted TLS clients
The list of TLS certificates trusted by a LXD server can be obtained with
`lxc config trust list`.

Clients can manually be added using `lxc config trust add <file>`,
removing the need for a shared trust password by letting an existing
administrator add the new client certificate directly to the trust store.

To revoke trust to a client its certificate can be removed with `lxc config
trust remove FINGERPRINT`.

## Password prompt with TLS authentication
To establish a new trust relationship when not already setup by the
administrator, a password must be set on the server and sent by the
client when adding itself.

A remote add operation should therefore go like this:

 1. Call GET /1.0
 2. If we're not in a PKI setup ask the user to confirm the fingerprint.
 3. Look at the dict we received back from the server. If "auth" is
    "untrusted", ask the user for the server's password and do a `POST` to
    `/1.0/certificates`, then call `/1.0` again to check that we're indeed
    trusted.
 4. Remote is now ready

## Failure scenarios
### Server certificate changes
This will typically happen in two cases:

 * The server was fully reinstalled and so changed certificate
 * The connection is being intercepted (MITM)

In such cases the client will refuse to connect to the server since the
certificate fringerprint will not match that in the config for this
remote.

It is then up to the user to contact the server administrator to check
if the certificate did in fact change. If it did, then the certificate
can be replaced by the new one or the remote be removed altogether and
re-added.

### Server trust relationship revoked
In this case, the server still uses the same certificate but all API
calls return a 403 with an error indicating that the client isn't
trusted.

This happens if another trusted client or the local server administrator
removed the trust entry on the server.

## Production setup
For production setup, it's recommended that `core.trust_password` is unset
after all clients have been added.  This prevents brute-force attacks trying to
guess the password.

Furthermore, `core.https_address` should be set to the single address where the
server should be available (rather than any address on the host), and firewall
rules should be set to only allow access to the LXD port from authorized
hosts/subnets.
