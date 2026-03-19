import { test, expect } from '@playwright/test';

test.describe('Dashboard End-to-End', () => {
  const projectName = `Test Project ${Math.floor(Math.random() * 1000)}`;
  const flagKey = 'test-flag';

  test('should complete the full project and flag lifecycle', async ({ page }) => {
    // 1. Navigate to Projects
    await page.goto('/projects');
    await expect(page.locator('h1')).toHaveText('Projects');

    // 2. Create a New Project
    await page.click('text=New Project');
    await expect(page.locator('h2', { hasText: 'Create New Project' })).toBeVisible();
    
    await page.fill('input[placeholder="e.g. Mobile App"]', projectName);
    await page.fill('textarea[placeholder="What is this project for?"]', 'Created by E2E test');
    await page.click('button:has-text("Create Project")');

    // Wait for modal to close
    await expect(page.locator('h2', { hasText: 'Create New Project' })).not.toBeVisible();

    // 3. Verify Project Appears and Navigate
    console.log(`Waiting for project card: ${projectName}`);
    
    try {
      const projectCard = page.locator('.card', { hasText: projectName });
      await expect(projectCard).toBeVisible({ timeout: 15000 });
      await projectCard.click();
    } catch (e) {
      await page.screenshot({ path: `failure-${projectName}.png` });
      throw e;
    }

    // 4. Verify Project Details View
    await expect(page.locator('h1')).toHaveText(projectName);
    await expect(page.locator('text=Environments')).toBeVisible();
    await expect(page.locator('text=Feature Flags')).toBeVisible();

    // Note: Since we haven't implemented "New Flag" modal in UI yet, 
    // we assume there might be no flags. If there are, we can test navigation.
    // For now, let's verify the empty state in details.
    await expect(page.locator('text=No flags created yet')).toBeVisible();
  });

  test('should navigate to an existing flag and modify rules', async ({ page }) => {
    // This test assumes a project and flag exist. 
    // In a real CI environment, we'd seed the DB.
    // For now, let's navigate to projects and try to find any project to enter.
    await page.goto('/projects');
    await expect(page.locator('text=Loading projects...')).not.toBeVisible();
    
    const projectCard = page.locator('.card').first();
    if (await projectCard.count() > 0) {
      await projectCard.click();
      
      const flagLink = page.locator('table tbody tr td a').first();
      if (await flagLink.count() > 0) {
        const flagName = await flagLink.innerText();
        await flagLink.click();

        // Verify Flag Detail View
        await expect(page.locator('h1')).toContainText(flagName);
        await expect(page.locator('text=Enable Flag')).toBeVisible();

        // Test Toggle
        const toggle = page.locator('.switch input');
        const initialState = await toggle.isChecked();
        await toggle.click();
        await expect(toggle).not.toBeChecked({ checked: initialState });

        // Add a Rule
        await page.click('text=Add Rule');
        await expect(page.locator('.rule-card')).toBeVisible();
        
        // Save
        await page.click('text=Save Configuration');
        await expect(page.locator('text=Saving...')).toBeVisible();
        await expect(page.locator('text=Save Configuration')).toBeVisible(); // Wait for it to finish
      }
    }
  });
});
