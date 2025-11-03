# TL Pokedex

## Endpoints    
1. Return basic Pokemon information.
```
GET /pokemon/<pokemon name>
```
Response: 
```
name: str
desc: str
habitat: str (? or enum)
is_legendary: bool
```

2. Return basic Pokemon information but with a ‘fun’ translation of the Pokemon description.
```
GET /pokemon/translated/<pokemon name>
```
Response schema is the same but we apply a transformation to the description.
Transformation = yoda if (pkmn.habitat == "cave" or pkmn.is_legendary) else "shakespeare"
If the transformation fails, we return the standard description.

Data should be fetched from:
- [PokéAPI](https://pokeapi.co/): most of what you need is under the pokemon-species API. The Pokemon’s description can be found under the flavor_text array. You can use any of the English descriptions.
- optionally translated with the [FunTranslations API](https://funtranslations.com)

## General idea
- Create services for getting pokemon data, translating descriptions.
- Create a module to cache data, we'll keep it in memory for this version, in production we'd use something like Redis.
- Maybe we can expose another endpoint that lists some of the recent viewed pokemon from other users to explore the Pokedex. If this were a real app we'd want to suggest other pokemon through more criteria like: evolutions, same type, same generation etc. For now though going by recently seen helps us to also limit traffic to external APIs since we'll likely have cache hits.
- we can also create rate limiters, mostly to the API we use so that we avoid being an irresponsible customer. This would come at the expense of latency to the user. 
- we can also keep track in memory of which requests are currently pending (e.g. for mewtwo), so that we don't start another request
 for something that soon will be in the cache
- we could also gather stats about the most requested pokemon, so that if we ever have a cold start we can start loading the most wanted items in the cache
- maybe we should normalize names somehow? In a real app I'd want to suggest corrections for misspellings


-- to parse --
● If you would have made a different design decision for production, then comment or
document it.
● We love high-value unit tests which test the right things!
● The task requirements are fairly trivial - we’re more interested in your design decisions,
code layout and approach (be prepared to explain this).

Please describe in the README.md:
● How to run it (don't assume anything is already installed)
● Anything you’d do differently for a production API
Bonus points for:
● Dockerfile
● Include your git history
Have fun, take your time and when you are done please send a link to your public Github repo to
your Talent Acquisition Partner!



