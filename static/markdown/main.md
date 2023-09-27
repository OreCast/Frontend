![Architecture](images/OreCastInfrastructure.png)

# OreCast architecture
The OreCast architecture is based on loosely coupled set of MicroServices:
- the frontend service to provide web UI interface to end-users
- the data discovery service to keep track of participated sites
- the meta-data service to keep track of meta-data information
- the data-management service to manage on-site data via S3 storage objects
- the data-bookkeeping service to keep provenance information about dataset
  processing
All of them are glued together by HTTP protocol and represent whole
infrustructure. For further details please refer to [implementation](docs/implementation.md)
details.

So far, the OreCast framework is work in progress, please refer to our
current list of [TODO tasks](docs/TODO.md).

We rely on many different technologies which we outline in
[references](docs/references.md) document.
