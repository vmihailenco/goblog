// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gorilla/schema fills a struct with form values.

The basic usage is really simple. Given this struct:

	type Person struct {
		Name  string
		Phone string
	}

...we can fill it passing a map to the Load() function:

	values := map[string][]string{
		"Name":  {"John"},
		"Phone": {"999-999-999"},
	}
	person := new(Person)
	schema.Load(person, values)

This is just a simple example and it doesn't make a lot of sense to create
the map manually. Typically it will come from a http.Request object and
will be of type url.Values: http.Request.Form or http.Request.MultipartForm.

The supported field types in the destination struct are:

- bool;

- float variants (float32, float64);

- int variants (int, int8, int16, int32, int64);

- string;

- uint variants (uint, uint8, uint16, uint32, uint64);

- structs;

- slices of any of the above types or maps with string keys and any of the
above types;

- types with one of the above underlying types.

- a pointer to any of the above types.

Non-supported types are simply ignored.

Nested structs are scanned recursivelly and the source keys must use dotted
notation for that. So for example, when filling the struct Person below:

	type Phone struct {
		Label  string
		Number string
	}

	type Person struct {
		Name  string
		Phone Phone
	}

...it will search for keys "Name", "Phone.Label" and "Phone.Number" in the
source map. Dotted names are needed to avoid name clashes. This means that
an HTML form to fill a Person struct must look like this:

	<form>
		<input type="text" name="Name">
		<input type="text" name="Phone.Label">
		<input type="text" name="Phone.Number">
	</form>

Single values are filled using the first value for a key from the source map.
Slices are filled using all values for a key from the source map. So to fill
a Person with multiple Phone values, like:

	type Person struct {
		Name   string
		Phones []Phone
	}

...an HTML form that accepts three Phone values would look like this:

	<form>
		<input type="text" name="Name">
		<input type="text" name="Phones.Label">
		<input type="text" name="Phones.Number">
		<input type="text" name="Phones.Label">
		<input type="text" name="Phones.Number">
		<input type="text" name="Phones.Label">
		<input type="text" name="Phones.Number">
	</form>

Maps can only have a string as key, and use the same dotted notation. So for
the struct:

	type Person struct {
		Name   string
		Scores map[string]int
	}

...we can define a form like:

	<form>
		<input type="text" name="Name">
		<input type="text" name="Scores.Math" value="7">
		<input type="text" name="Scores.RocketScience" value="1">
		<input type="text" name="Scores.Go" value="9">
	</form>

...and the resulting Scores map will be:

	map[string]int{
		"Math":          7,
		"RocketScience": 1,
		"Go":            9,
	}

As you see, not everybody is good at rocket science!
*/
package schema
