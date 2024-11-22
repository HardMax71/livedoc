import { Component, createSignal } from 'solid-js';
import { useNavigate } from '@solidjs/router';
import Input from '../../components/common/Input';
import Button from '../../components/common/Button';
import { documentStore } from '@/stores/document.ts';

const NewDocument: Component = () => {
    const navigate = useNavigate();
    const [title, setTitle] = createSignal('');

    const handleSubmit = async (e: Event) => {
        e.preventDefault();
        const doc = await documentStore.createDocument(title());
        if (doc) {
            navigate(`/documents/${doc.id}`);
        }
    };

    return (
        <div class="container">
            <div class="columns is-centered">
                <div class="column is-6">
                    <div class="box mt-6">
                        <h1 class="title">New Document</h1>
                        <form onSubmit={handleSubmit}>
                            <Input
                                type="text"
                                label="Title"
                                value={title()}
                                onInput={(e) => setTitle(e.currentTarget.value)}
                                required
                            />
                            <div class="field">
                                <Button
                                    type="submit"
                                    isLoading={documentStore.state.isLoading}
                                >
                                    Create
                                </Button>
                                <Button
                                    type="button"
                                    variant="ghost"
                                    onClick={() => navigate('/documents')}
                                >
                                    Cancel
                                </Button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default NewDocument;