#!/bin/sh

tcli utils echo -format 'range(1;2) | {body: {id:.,name:.|tostring,photourls:[.|tostring]}}' \
  | tcli petstore pet addPet -format '{petId:.id}' \
  | tcli petstore pet getPetById -format '{petId:.id,api_key:.id}' \
  | tcli petstore pet deletePet -format '{petId:.message}' \
  | tcli petstore pet getPetById -status_code 404 -v
