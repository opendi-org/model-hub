import React, { useState } from 'react';
import { Container, TextField, IconButton, List, ListItem, ListItemText, Box } from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import AddIcon from '@mui/icons-material/Add';
import API_URL from '../config';
import { FormControl, InputLabel, MenuItem, Select } from '@mui/material';
import ModelMinicard from '../components/ModelMinicard'
import { useSearchParams } from "react-router-dom";
import { useEffect } from 'react';


const SearchPage = () => {
    const [searchParams, setSearchParams] = useSearchParams(window.location.search);
    const [searchTerm, setSearchTerm] = useState(searchParams.get('term') ?? '');
    const [results, setResults] = useState([]);
    const [searchType, setSearchType] = React.useState(searchParams.get('type') ?? 'model');


    const handleSearchChange = (event) => {
        setSearchTerm(event.target.value);
    };

    const handleSearch = () => {
        fetch(`${API_URL}/v0/models/search/${searchType}/${searchTerm}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(data => {
                setResults(data)
            }).catch(error => console.error('There was an error fetching search results:', error));
    };

    useEffect(() => {
        if (searchTerm !== '') {
            handleSearch()
        }
    }, []);


    const handleChange = (event) => {
        setSearchType(event.target.value);
    };

    return (
        <Container>
            <Box display="flex" justifyContent="center" alignItems="center" mt={2}>
                <TextField
                    variant="outlined"
                    placeholder="Search"
                    value={searchTerm}
                    onChange={handleSearchChange}
                    InputProps={{
                        startAdornment: (
                            <IconButton onClick={handleSearch}>
                                <SearchIcon />
                            </IconButton>
                        ),
                    }}
                    onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                            handleSearch()
                        }
                    }}
                />
                <FormControl>
                <InputLabel>Filter</InputLabel>
                <Select
                    labelId="select-search"
                    id="select-search"
                    value={searchType}
                    label="Filter"
                    onChange={handleChange}
                >
                    <MenuItem value={"model"}>Model Name</MenuItem>
                    <MenuItem value={"user"}>Creator Name</MenuItem>
                </Select>
                </FormControl>
                            </Box>
            <List>
                {results.map((result, index) => (
                    <ModelMinicard key={result.meta.uuid} name={result.meta.name} id = {result.meta.uuid} author={result.meta.creator.username} summary={result.meta.summary}>
                    </ModelMinicard>
                ))}
            </List>
        </Container>
    );
};

export default SearchPage