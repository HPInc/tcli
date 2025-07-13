#!/bin/sh
# increases test input to 10 (1..10)
# runs all api calls in parallel

tcli utils echo -format 'range(1;11) | {body: {id:.,name:.|tostring,photourls:[.|tostring]}}' \
  | tcli petstore pet addPet -format '{petId:.id}' -parallel \
  | tcli petstore pet getPetById -format '{petId:.id,api_key:.id}' -parallel \
  | tcli petstore pet deletePet -format '{petId:.message}' -parallel \
  | tcli petstore pet getPetById -status_code 404 -parallel
