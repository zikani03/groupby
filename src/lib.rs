use std::collections::BTreeMap;
use std::cmp::Ord;
use std::clone::Clone;

/// GroupBy Trait
/// Groups items of a specified type `T` by a function `F`.
/// The groups are inserted into a map that is indexed by the output
/// of the function `f`.
pub trait GroupBy<T,F,G> {
	fn group_by(&self, f: F) -> BTreeMap<G, Vec<T>>
		where T : Ord + Clone, 
			  G : Ord,
			  F : for <'a> Fn(&'a T) -> G;
}


/// Implementation of `GroupBy` for vectors.
/// Groups values in a vector by function `f`
impl<T,F,G> GroupBy<T,F,G> for Vec<T> {

	/// Group values of a vector into a BTreeMap
	fn group_by(&self, f: F) -> BTreeMap<G, Vec<T>>
		where T : Ord + Clone, 
			  G : Ord,
			  F : for<'a> Fn(&'a T) -> G {

		let mut group = BTreeMap::<G, Vec<T>>::new();
		
		for value in self {
			let key = f(value);
			
			if group.contains_key(&key) {
				if let Some(vals) = group.get_mut(&key){
					vals.push(value.clone());
				}
			} else {
				group.insert(key, vec![value.clone()]);
			}
		}

		let res = group;
		res
	}
}

/// Implementation of `GroupBy` for BTreeMaps.
/// Groups values in a BTreeMap by function `f`
impl<T,F,G> GroupBy<T,F,G> for BTreeMap<G, T> {

	/// Group values of a vector into a BTreeMap
	fn group_by(&self, f: F) -> BTreeMap<G, BTreeMap<G, T>>
		where T : Ord + Clone, 
			  G : Ord,
			  F : for<'a> Fn(&'a T) -> G {

		let mut group = BTreeMap::<G, BTreeMap<G, T>>::new();
		
		for value in self {
			let key = f(value);
			
			if group.contains_key(&key) {
				if let Some(vals) = group.get_mut(&key){
					vals.push(value.clone());
				}
			} else {
				group.insert(key, vec![value.clone()]);
			}
		}

		let res = group;
		res
	}
}