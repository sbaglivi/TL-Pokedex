# TL Pokedex

## Description
A small API written in Go. It offers data about pokemons searched through their name, optionally with a fun twist on the pokemon description.  
All responses are returned as JSON.  
It is powered by the [PokéAPI](https://pokeapi.co) and by the [Funtranslations API](https://api.funtranslations.com/).

## How to run the API
To run the application:
- install **Git**: 
  full instructions are available [here](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git?pStoreID=newegg/1000%270%27A=0%27[0])
- choose the folder where you'd like to install the project, open a terminal in it and then clone the project `git clone https://github.com/sbaglivi/TL-Pokedex`
- enter the folder you just created `cd TL-Pokedex`

Now, if you'd like to run the application using Docker:
- install **Docker**:
  if you use linux see [here](https://docs.docker.com/engine/install/) for instructions;  
  otherwise look [here](https://docs.docker.com/desktop/)
- run the application: if you have Make installed, you can run `make docker-run`, otherwise you can use the full command `docker build -t tl-pokedex --platform linux/amd64 . && docker run -it --rm --platform linux/amd64 -p 3000:3000 tl-pokedex`.  

If instead you'd like to run it outside a container:
- install **Go**: 
  download the appropriate version for your OS [here](https://go.dev/dl/) and then follow the instructions (specific to your OS) [here](https://go.dev/doc/install)
- run the application either in development mode: `go run main.go` or build it and then run it `go build -o bin/pokedex && ./bin/pokedex`

By default the app will be listening on port 3000.

## Usage
Once the web server is up and running, there should be 2 endpoints available:
- `GET http://localhost:3000/pokemon/{pokemon_name}`  
Searches for a pokemon named `{pokemon_name}`   
If it doesn't find it, it responds with a status code of 404, and a response body `{"error": "not found"}`  
If an unforeseen error happens, it responds with: 500, `{"error": "internal server error"}`  
If everything goes well, an example response looks like this (status code = 200):
```json
{
  "pokemon": {
    "is_legendary": false,
    "name": "espeon",
    "habitat": "urban",
    "desc": "It uses the fine hair that covers its body to sense air currents and predict its ene­mies actions."
  }
}
```
- `GET http://localhost:3000/pokemon/translated/{pokemon_name}` 
Searches for a pokemon named `{pokemon_name}` but tries to use the Funtranslations API to modify its description.  
If everything goes well, the response is exactly like the one above (except for the different description content).  
In case the Pokemon search encounters an error, the same errors from the previous endpoint might be returned (404, 500).  
In case the translation encounters a problem - most often because of rate limits - it returns, in addition to the pokemon info, a top-level key in the response `warnings` that informs the user that the translation failed (e.g. `"warnings": ["translation failed"]`).

## Implementation and possible improvements
The protagonist of the API is the PokemonService (package pokemon). This service depends on a translation service (injected at initialization) and is responsible for fetching data about Pokemons (from the PokeAPI or from a cache) and optionally translate their description.  
A different solution could've been to create a higher level component that used both the PokemonService and the TranslationService to fulfill the API needs, to avoid giving the responsibility of translations to the PokemonService.  
For this particular use case, where the only consumer of the TranslationService is the PokemonService, I thought it wasn't necessary.  

Both services utilize a LRU cache to avoid making multiple requests for the same pokemons when possible. A LFU cache could also have been used, and probably even better since I imagine the popularity of a few number of pokemons is vastly superior to the rest.  
Again, for this particular use case I don't think it matters much, the data we need to handle is so small that we could probably cache all the existing pokemons without ever needing to worry about eviction policies.

Some changes that I'd implement if this was a real application:
- add ways to explore pokemons: an endpoint for most frequently searched, add data about the evolutions of the current search, pokemon of the same type, etc.
- use an external cache, so that if we need to restart the application we won't start from scratch, and if we're running multiple instances of it, we can share data instead of having different copies of the cache
- add authentication, and rate limiting or a paid plan (or both) so that we can either pay for use of the Funtranslation API or prevent any single user from consuming all free requests
- suggest corrections for misspelled pokemon names


