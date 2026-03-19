import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import App from './App';

describe('Dashboard Component', () => {
  it('renders the sidebar and main header', async () => {
    render(<App />);
    
    expect(screen.getByText('Solid Fortnight')).toBeInTheDocument();
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
  });

  it('opens the "Create New Project" modal when clicking the button', async () => {
    render(<App />);
    
    // Click on Projects in sidebar
    fireEvent.click(screen.getByText('Projects'));
    
    // Wait for header
    expect(await screen.findByRole('heading', { name: /Projects/i })).toBeInTheDocument();

    // Click New Project
    const newProjectBtn = screen.getByText(/New Project/i);
    fireEvent.click(newProjectBtn);

    // Verify modal appears
    expect(screen.getByText('Create New Project')).toBeInTheDocument();
    expect(screen.getByLabelText('Project Name')).toBeInTheDocument();
  });
});
